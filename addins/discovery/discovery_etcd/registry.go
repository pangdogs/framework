/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package discovery_etcd

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"slices"
	"time"

	"git.golaxy.org/core"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/addins/discovery"
	"git.golaxy.org/framework/addins/log"
	"github.com/elliotchance/pie/v2"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

func newEtcdRegistry(settings ...option.Setting[EtcdRegistryOptions]) discovery.IRegistry {
	return &_EtcdRegistry{
		options: option.New(With.Default(), settings...),
	}
}

type _EtcdRegistry struct {
	svcCtx    service.Context
	ctx       context.Context
	terminate context.CancelFunc
	barrier   generic.Barrier
	options   EtcdRegistryOptions
	client    *etcdv3.Client
}

// Init 初始化插件
func (r *_EtcdRegistry) Init(svcCtx service.Context) {
	log.L(svcCtx).Info("initializing add-in", zap.String("name", AddIn.Name))

	r.svcCtx = svcCtx
	r.ctx, r.terminate = context.WithCancel(context.Background())

	if r.options.EtcdClient == nil {
		cli, err := etcdv3.New(r.configure())
		if err != nil {
			log.L(svcCtx).Panic("new etcd client failed", log.JSON("config", r.configure()), zap.Error(err))
		}
		r.client = cli
	} else {
		r.client = r.options.EtcdClient
	}

	for _, ep := range r.client.Endpoints() {
		func() {
			ctx, cancel := context.WithTimeout(svcCtx, 3*time.Second)
			defer cancel()

			if _, err := r.client.Status(ctx, ep); err != nil {
				log.L(svcCtx).Panic("status etcd failed", zap.Any("endpoint", ep), zap.Error(err))
			}
		}()
	}
}

// Shut 关闭插件
func (r *_EtcdRegistry) Shut(svcCtx service.Context) {
	log.L(svcCtx).Info("shutting down add-in", zap.String("name", AddIn.Name))

	r.terminate()
	r.barrier.Close()
	r.barrier.Wait()

	if r.options.EtcdClient == nil {
		if r.client != nil {
			r.client.Close()
		}
	}
}

// RegisterNode 注册服务节点
func (r *_EtcdRegistry) RegisterNode(ctx context.Context, serviceName string, node *discovery.Node, ttl time.Duration) (discovery.IRegistration, error) {
	if serviceName == "" {
		return nil, fmt.Errorf("registry: %w serviceName is empty", core.ErrArgs)
	}
	if node == nil {
		return nil, fmt.Errorf("registry: %w node is nil", core.ErrArgs)
	}
	if node.Id == "" {
		return nil, fmt.Errorf("registry: %w node.id is empty", core.ErrArgs)
	}
	return r.registerNode(ctx, serviceName, node, ttl)
}

// Get 查询服务
func (r *_EtcdRegistry) Get(ctx context.Context, serviceName string) (*discovery.Service, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if serviceName == "" {
		return nil, discovery.ErrRegistrationNotFound
	}

	serviceKey := r.newServiceKey(serviceName) + "/"

	rsp, err := r.client.Get(ctx, serviceKey,
		etcdv3.WithPrefix(),
		etcdv3.WithSort(etcdv3.SortByModRevision, etcdv3.SortDescend))
	if err != nil {
		log.L(r.svcCtx).Error("get etcd key failed", zap.String("key", serviceKey), zap.Error(err))
		return nil, fmt.Errorf("registry: %w", err)
	}

	if len(rsp.Kvs) <= 0 {
		return nil, discovery.ErrRegistrationNotFound
	}

	service := &discovery.Service{
		Name:     serviceName,
		Nodes:    make([]discovery.Node, 0, len(rsp.Kvs)),
		Revision: rsp.Header.Revision,
	}

	for _, kv := range rsp.Kvs {
		serviceNode, err := decodeService(kv.Value)
		if err != nil {
			log.L(r.svcCtx).Error("decode service failed", zap.ByteString("key", kv.Key), zap.Error(err))
			continue
		}

		if len(serviceNode.Nodes) <= 0 {
			log.L(r.svcCtx).Error("decode service failed", zap.ByteString("key", kv.Key), zap.Error(errors.New("nodes is empty")))
			continue
		}

		service.Nodes = append(service.Nodes, serviceNode.Nodes...)
	}

	return service, nil
}

// GetNode 查询服务节点
func (r *_EtcdRegistry) GetNode(ctx context.Context, serviceName string, nodeId uid.Id) (*discovery.Service, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if serviceName == "" || nodeId == "" {
		return nil, discovery.ErrRegistrationNotFound
	}

	nodeKey := r.newNodeKey(serviceName, nodeId)

	rsp, err := r.client.Get(ctx, nodeKey)
	if err != nil {
		log.L(r.svcCtx).Error("get etcd key failed", zap.String("key", nodeKey), zap.Error(err))
		return nil, fmt.Errorf("registry: %w", err)
	}

	if len(rsp.Kvs) <= 0 {
		return nil, discovery.ErrRegistrationNotFound
	}

	serviceNode, err := decodeService(rsp.Kvs[0].Value)
	if err != nil {
		return nil, fmt.Errorf("registry: %w", err)
	}
	serviceNode.Revision = rsp.Header.Revision

	return serviceNode, nil
}

// List 查询所有服务
func (r *_EtcdRegistry) List(ctx context.Context) ([]*discovery.Service, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	rsp, err := r.client.Get(ctx, r.options.KeyPrefix,
		etcdv3.WithPrefix(),
		etcdv3.WithSort(etcdv3.SortByModRevision, etcdv3.SortDescend))
	if err != nil {
		log.L(r.svcCtx).Error("get etcd key failed", zap.String("key", r.options.KeyPrefix), zap.Error(err))
		return nil, fmt.Errorf("registry: %w", err)
	}

	if len(rsp.Kvs) <= 0 {
		return nil, nil
	}

	var services []*discovery.Service

	for _, kv := range rsp.Kvs {
		serviceNode, err := decodeService(kv.Value)
		if err != nil {
			log.L(r.svcCtx).Error("decode service failed", zap.ByteString("key", kv.Key), zap.Error(err))
			continue
		}

		if len(serviceNode.Nodes) <= 0 {
			log.L(r.svcCtx).Error("decode service failed", zap.ByteString("key", kv.Key), zap.Error(errors.New("nodes is empty")))
			continue
		}

		serviceNode.Revision = rsp.Header.Revision

		idx := slices.IndexFunc(services, func(service *discovery.Service) bool {
			return service.Name == serviceNode.Name
		})
		if idx < 0 {
			services = append(services, serviceNode)
			continue
		}

		service := services[idx]

		if service.Revision < serviceNode.Revision {
			service.Revision = serviceNode.Revision
		}

		service.Nodes = append(service.Nodes, serviceNode.Nodes...)
	}

	return services, nil
}

// WatchEvent 观察服务变化事件流
func (r *_EtcdRegistry) WatchEvent(ctx context.Context, pattern string, revision ...int64) (<-chan discovery.Event, error) {
	eventChan, _, err := r.addWatcher(ctx, pattern, nil, pie.First(revision))
	if err != nil {
		return nil, err
	}
	return eventChan, nil
}

// WatchHandler 观察服务变化事件回调
func (r *_EtcdRegistry) WatchHandler(ctx context.Context, pattern string, handler discovery.EventHandler, revision ...int64) (async.Future, error) {
	if handler == nil {
		return async.Future{}, fmt.Errorf("registry: %w: handler is nil", core.ErrArgs)
	}
	_, stopped, err := r.addWatcher(ctx, pattern, handler, pie.First(revision))
	if err != nil {
		return async.Future{}, err
	}
	return stopped, err
}

func (r *_EtcdRegistry) configure() etcdv3.Config {
	if r.options.EtcdConfig != nil {
		return *r.options.EtcdConfig
	}

	config := etcdv3.Config{
		Endpoints:   r.options.CustomAddresses,
		Username:    r.options.CustomUsername,
		Password:    r.options.CustomPassword,
		DialTimeout: 3 * time.Second,
	}

	if r.options.CustomTLSConfig != nil {
		tlsConfig := r.options.CustomTLSConfig
		if tlsConfig == nil {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		config.TLS = tlsConfig
	}

	return config
}

func (r *_EtcdRegistry) newNodeKey(service string, nodeId uid.Id) string {
	return path.Join(r.options.KeyPrefix, service, nodeId.String())
}

func (r *_EtcdRegistry) newServiceKey(service string) string {
	return path.Join(r.options.KeyPrefix, service)
}

func encodeService(s *discovery.Service) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func decodeService(ds []byte) (*discovery.Service, error) {
	var s discovery.Service
	if err := json.Unmarshal(ds, &s); err != nil {
		return nil, err
	}
	return &s, nil
}
