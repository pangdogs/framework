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

package etcd_discovery

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/addins/discovery"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/utils/concurrent"
	hash "github.com/mitchellh/hashstructure/v2"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"math"
	"path"
	"slices"
	"strings"
	"time"
)

// NewRegistry 创建registry插件，可以配合registry cache将数据缓存本地，提高查询效率
func NewRegistry(settings ...option.Setting[RegistryOptions]) discovery.IRegistry {
	return &_Registry{
		options: option.Make(With.Default(), settings...),
	}
}

type _Registration struct {
	ctx       context.Context
	terminate context.CancelFunc
	hash      uint64
	leaseId   etcdv3.LeaseID
	revision  int64
}

type _Registry struct {
	svcCtx        service.Context
	options       RegistryOptions
	client        *etcdv3.Client
	registrations *concurrent.Cache[string, *_Registration]
}

// Init 初始化插件
func (r *_Registry) Init(svcCtx service.Context) {
	log.Infof(svcCtx, "init addin %q", self.Name)

	r.svcCtx = svcCtx

	if r.options.EtcdClient == nil {
		cli, err := etcdv3.New(r.configure())
		if err != nil {
			log.Panicf(r.svcCtx, "new etcd client failed, %s", err)
		}
		r.client = cli
	} else {
		r.client = r.options.EtcdClient
	}

	for _, ep := range r.client.Endpoints() {
		func() {
			ctx, cancel := context.WithTimeout(r.svcCtx, 3*time.Second)
			defer cancel()

			if _, err := r.client.Status(ctx, ep); err != nil {
				log.Panicf(r.svcCtx, "status etcd %q failed, %s", ep, err)
			}
		}()
	}

	r.registrations = concurrent.NewCache[string, *_Registration]()
	r.registrations.OnDel(func(nodePath string, registration *_Registration) { registration.terminate() })
}

// Shut 关闭插件
func (r *_Registry) Shut(svcCtx service.Context) {
	log.Infof(svcCtx, "shut addin %q", self.Name)

	if r.options.EtcdClient == nil {
		if r.client != nil {
			r.client.Close()
		}
	}
}

// Register 注册服务
func (r *_Registry) Register(ctx context.Context, service *discovery.Service, ttl time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if service == nil {
		return fmt.Errorf("registry: %w: serivce is nil", core.ErrArgs)
	}

	if len(service.Nodes) <= 0 {
		return errors.New("registry: require at least one node")
	}

	var errs []error

	for i := range service.Nodes {
		node := &service.Nodes[i]

		if err := r.registerNode(ctx, service.Name, node, ttl); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", node.Id, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("registry: %w", errors.Join(errs...))
	}

	return nil
}

// Deregister 取消注册服务
func (r *_Registry) Deregister(ctx context.Context, service *discovery.Service) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if service == nil {
		return fmt.Errorf("registry: %w: serivce is nil", core.ErrArgs)
	}

	if len(service.Nodes) <= 0 {
		return fmt.Errorf("registry: require at least one node")
	}

	var errs []error

	for i := range service.Nodes {
		node := &service.Nodes[i]

		if err := r.deregisterNode(ctx, service.Name, node); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", node.Id, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("registry: %w", errors.Join(errs...))
	}

	return nil
}

// RefreshTTL 刷新所有服务TTL
func (r *_Registry) RefreshTTL(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	snapshot := r.registrations.Snapshot()
	var errs []error

	for _, kv := range snapshot {
		_, err := r.client.KeepAliveOnce(ctx, kv.V.leaseId)
		if err != nil {
			errs = append(errs, fmt.Errorf("keeplive %q failed, %w", kv.K, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("registry: %w", errors.Join(errs...))
	}

	return nil
}

// GetServiceNode 查询服务节点
func (r *_Registry) GetServiceNode(ctx context.Context, serviceName string, nodeId uid.Id) (*discovery.Service, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if serviceName == "" || nodeId == "" {
		return nil, discovery.ErrNotFound
	}

	rsp, err := r.client.Get(ctx, getNodePath(r.options.KeyPrefix, serviceName, nodeId))
	if err != nil {
		return nil, fmt.Errorf("registry: %w", err)
	}

	if len(rsp.Kvs) <= 0 {
		return nil, discovery.ErrNotFound
	}

	return decodeService(rsp.Kvs[0].Value)
}

// GetService 查询服务
func (r *_Registry) GetService(ctx context.Context, serviceName string) (*discovery.Service, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if serviceName == "" {
		return nil, discovery.ErrNotFound
	}

	rsp, err := r.client.Get(ctx, getServicePath(r.options.KeyPrefix, serviceName),
		etcdv3.WithPrefix(),
		etcdv3.WithSort(etcdv3.SortByModRevision, etcdv3.SortDescend))
	if err != nil {
		return nil, fmt.Errorf("registry: %w", err)
	}

	if len(rsp.Kvs) <= 0 {
		return nil, discovery.ErrNotFound
	}

	service := &discovery.Service{
		Name:     serviceName,
		Nodes:    make([]discovery.Node, 0, len(rsp.Kvs)),
		Revision: rsp.Header.Revision,
	}

	for _, kv := range rsp.Kvs {
		serviceNode, err := decodeService(kv.Value)
		if err != nil {
			log.Errorf(r.svcCtx, "decode service %q failed, %s", kv.Value, err)
			continue
		}

		if len(serviceNode.Nodes) <= 0 {
			log.Errorf(r.svcCtx, "decode service %q failed, nodes is empty", kv.Value)
			continue
		}

		service.Nodes = append(service.Nodes, serviceNode.Nodes...)
	}

	return service, nil
}

// ListServices 查询所有服务
func (r *_Registry) ListServices(ctx context.Context) ([]discovery.Service, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	rsp, err := r.client.Get(ctx, r.options.KeyPrefix,
		etcdv3.WithPrefix(),
		etcdv3.WithSort(etcdv3.SortByModRevision, etcdv3.SortDescend))
	if err != nil {
		return nil, fmt.Errorf("registry: %w", err)
	}

	if len(rsp.Kvs) <= 0 {
		return nil, nil
	}

	var services []discovery.Service

	for _, kv := range rsp.Kvs {
		serviceNode, err := decodeService(kv.Value)
		if err != nil {
			log.Errorf(r.svcCtx, "decode service %q failed, %s", kv.Value, err)
			continue
		}

		if len(serviceNode.Nodes) <= 0 {
			log.Errorf(r.svcCtx, "decode service %q failed, nodes is empty", kv.Value)
			continue
		}

		serviceNode.Revision = rsp.Header.Revision

		idx := slices.IndexFunc(services, func(service discovery.Service) bool {
			return service.Name == serviceNode.Name
		})
		if idx < 0 {
			services = append(services, *serviceNode)
			continue
		}

		service := &services[idx]

		if service.Revision < serviceNode.Revision {
			service.Revision = serviceNode.Revision
		}

		service.Nodes = append(service.Nodes, serviceNode.Nodes...)
	}

	return services, nil
}

// Watch 监听服务变化
func (r *_Registry) Watch(ctx context.Context, pattern string, revision ...int64) (discovery.IWatcher, error) {
	return r.newWatcher(ctx, pattern, revision...)
}

func (r *_Registry) configure() etcdv3.Config {
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

func (r *_Registry) registerNode(ctx context.Context, serviceName string, node *discovery.Node, ttl time.Duration) error {
	if serviceName == "" {
		return errors.New("service name can't empty")
	}

	if node.Id == "" {
		return errors.New("service node id can't empty")
	}

	ttl = max(ttl, r.options.TTL)

	hv, err := hash.Hash(node, hash.FormatV2, nil)
	if err != nil {
		return err
	}

	nodePath := getNodePath(r.options.KeyPrefix, serviceName, node.Id)
	var leaseId etcdv3.LeaseID

	registration, ok := r.registrations.Get(nodePath)
	if ok {
		_, err = r.client.KeepAliveOnce(ctx, registration.leaseId)
		if !errors.Is(err, rpctypes.ErrLeaseNotFound) {
			return err
		}
		if err == nil {
			if registration.hash == hv {
				log.Debugf(r.svcCtx, "service %q node %q unchanged, keep alive lease", serviceName, node.Id)
				return nil
			}
			leaseId = registration.leaseId
		}
	}

	servNode := &discovery.Service{
		Name:  serviceName,
		Nodes: []discovery.Node{*node},
	}
	servNodeData := encodeService(servNode)

	var revision int64

	if leaseId != etcdv3.NoLease {
		rsp, err := r.client.Put(ctx, nodePath, servNodeData, etcdv3.WithLease(leaseId))
		if err != nil {
			return err
		}
		revision = rsp.Header.Revision

	} else {
		lgr, err := r.client.Grant(ctx, int64(math.Ceil(ttl.Seconds())))
		if err != nil {
			return err
		}
		leaseId = lgr.ID

		rsp, err := r.client.Txn(ctx).
			If(etcdv3.Compare(etcdv3.Version(nodePath), "=", 0)).
			Then(etcdv3.OpPut(nodePath, servNodeData, etcdv3.WithLease(leaseId))).
			Commit()
		if err != nil {
			return err
		}

		if !rsp.Succeeded {
			return fmt.Errorf("service %q node %q already existed", serviceName, node.Id)
		}
		revision = rsp.Header.Revision
	}

	registration = &_Registration{
		hash:     hv,
		leaseId:  leaseId,
		revision: revision,
	}
	registration.ctx, registration.terminate = context.WithCancel(r.svcCtx)

	existed := r.registrations.Set(nodePath, registration, registration.revision, 0)
	if existed != registration {
		return nil
	}

	if r.options.AutoRefreshTTL {
		rspChan, err := r.client.KeepAlive(registration.ctx, leaseId)
		if err != nil {
			return err
		}
		go func() {
			for range rspChan {
				log.Debugf(r.svcCtx, "refresh service %q node %q ttl success", servNode.Name, node.Id)
			}
		}()
	}

	log.Debugf(r.svcCtx, "register service %q node %q success", servNode.Name, node.Id)
	return nil
}

func (r *_Registry) deregisterNode(ctx context.Context, serviceName string, node *discovery.Node) error {
	nodePath := getNodePath(r.options.KeyPrefix, serviceName, node.Id)

	registration, ok := r.registrations.Get(nodePath)
	if !ok {
		return nil
	}
	r.registrations.Del(nodePath, registration.revision+1)

	_, err := r.client.Revoke(ctx, registration.leaseId)
	if err != nil {
		return err
	}

	log.Debugf(r.svcCtx, "deregister service %q node %q success", serviceName, node.Id)
	return nil
}

func encodeService(s *discovery.Service) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func decodeService(ds []byte) (*discovery.Service, error) {
	var s *discovery.Service

	if err := json.Unmarshal(ds, &s); err != nil {
		return nil, err
	}

	return s, nil
}

func getNodePath(prefix, s string, id uid.Id) string {
	service := strings.ReplaceAll(s, "/", "-")
	node := strings.ReplaceAll(id.String(), "/", "-")
	return path.Join(prefix, service, node)
}

func getServicePath(prefix, s string) string {
	service := strings.ReplaceAll(s, "/", "-")
	return path.Join(prefix, service) + "/"
}
