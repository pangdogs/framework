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

package redis_discovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/core/utils/types"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/plugins/discovery"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/utils/concurrent"
	hash "github.com/mitchellh/hashstructure/v2"
	"github.com/redis/go-redis/v9"
	"slices"
	"sort"
	"strings"
	"time"
)

// NewRegistry 创建registry插件，可以配合registry cache将数据缓存本地，提高查询效率
func NewRegistry(settings ...option.Setting[RegistryOptions]) discovery.IRegistry {
	return &_Registry{
		options: option.Make(With.Default(), settings...),
	}
}

type _Register struct {
	hash     uint64
	ttl      time.Duration
	revision int64
}

type _Registry struct {
	servCtx   service.Context
	options   RegistryOptions
	client    *redis.Client
	registers *concurrent.Cache[string, *_Register]
}

// InitSP 初始化服务插件
func (r *_Registry) InitSP(ctx service.Context) {
	log.Infof(ctx, "init plugin %q", self.Name)

	r.servCtx = ctx

	if r.options.RedisClient == nil {
		r.client = redis.NewClient(r.configure())
	} else {
		r.client = r.options.RedisClient
	}

	_, err := r.client.Ping(r.servCtx).Result()
	if err != nil {
		log.Panicf(r.servCtx, "ping redis %q failed, %v", r.client, err)
	}

	_, err = r.client.ConfigSet(r.servCtx, "notify-keyspace-events", "KEA").Result()
	if err != nil {
		log.Panicf(r.servCtx, "redis %q enable notify-keyspace-events failed, %v", r.client, err)
	}

	r.registers = concurrent.NewCache[string, *_Register]()
}

// ShutSP 关闭服务插件
func (r *_Registry) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut plugin %q", self.Name)

	if r.options.RedisClient == nil {
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
		return errors.New("registry: require at least one node")
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

	snapshot := r.registers.Snapshot()
	var errs []error

	for _, kv := range snapshot {
		_, err := r.client.Expire(ctx, kv.K, kv.V.ttl).Result()
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

	nodeVal, err := r.client.Get(ctx, getNodePath(r.options.KeyPrefix, serviceName, nodeId)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, discovery.ErrNotFound
		}
		return nil, fmt.Errorf("registry: %w", err)
	}

	return decodeService(nodeVal)
}

// GetService 查询服务
func (r *_Registry) GetService(ctx context.Context, serviceName string) (*discovery.Service, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if serviceName == "" {
		return nil, discovery.ErrNotFound
	}

	nodeKeys, err := r.client.Keys(ctx, getServicePath(r.options.KeyPrefix, serviceName)).Result()
	if err != nil {
		return nil, fmt.Errorf("registry: %w", err)
	}

	if len(nodeKeys) <= 0 {
		return nil, discovery.ErrNotFound
	}

	nodeVals, err := r.client.MGet(ctx, nodeKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("registry: %w", err)
	}

	serviceList := make([]*discovery.Service, 0, len(nodeVals))

	for _, v := range nodeVals {
		service, err := decodeService(types.String2Bytes(v.(string)))
		if err != nil {
			log.Errorf(r.servCtx, "decode service %q failed, %s", v, err)
			continue
		}

		if len(service.Nodes) <= 0 {
			log.Errorf(r.servCtx, "decode service %q failed, nodes is empty", v)
			continue
		}

		serviceList = append(serviceList, service)
	}

	sort.Slice(serviceList, func(i, j int) bool {
		return serviceList[i].Revision > serviceList[j].Revision
	})

	service := &discovery.Service{
		Name:  serviceName,
		Nodes: make([]discovery.Node, 0, len(serviceList)),
	}

	if len(serviceList) > 0 {
		service.Revision = serviceList[0].Revision
	}

	for i := range serviceList {
		service.Nodes = append(service.Nodes, serviceList[i].Nodes...)
	}

	return service, nil
}

// ListServices 查询所有服务
func (r *_Registry) ListServices(ctx context.Context) ([]discovery.Service, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	nodeKeys, err := r.client.Keys(ctx, r.options.KeyPrefix+"*").Result()
	if err != nil {
		return nil, fmt.Errorf("registry: %w", err)
	}

	if len(nodeKeys) <= 0 {
		return nil, nil
	}

	nodeVals, err := r.client.MGet(ctx, nodeKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("registry: %w", err)
	}

	services := make([]*discovery.Service, 0, len(nodeVals))

	for _, v := range nodeVals {
		service, err := decodeService(types.String2Bytes(v.(string)))
		if err != nil {
			log.Errorf(r.servCtx, "decode service %q failed, %s", v, err)
			continue
		}

		if len(service.Nodes) <= 0 {
			log.Errorf(r.servCtx, "decode service %q failed, nodes is empty", v)
			continue
		}

		services = append(services, service)
	}

	sort.Slice(services, func(i, j int) bool {
		return services[i].Revision > services[j].Revision
	})

	var rets []discovery.Service

	for i := range services {
		service := services[i]

		idx := slices.IndexFunc(rets, func(ret discovery.Service) bool {
			return ret.Name == service.Name
		})
		if idx < 0 {
			rets = append(rets, *service)
			continue
		}

		ret := &rets[idx]

		if ret.Revision < service.Revision {
			ret.Revision = service.Revision
		}

		ret.Nodes = append(ret.Nodes, service.Nodes...)
	}

	return rets, nil
}

// Watch 监听服务变化
func (r *_Registry) Watch(ctx context.Context, pattern string, revision ...int64) (discovery.IWatcher, error) {
	return r.newWatcher(ctx, pattern)
}

func (r *_Registry) configure() *redis.Options {
	if r.options.RedisConfig != nil {
		return r.options.RedisConfig
	}

	if r.options.RedisURL != "" {
		conf, err := redis.ParseURL(r.options.RedisURL)
		if err != nil {
			log.Panicf(r.servCtx, "parse redis url %q failed, %s", r.options.RedisURL, err)
		}
		return conf
	}

	conf := &redis.Options{}
	conf.Username = r.options.CustomUsername
	conf.Password = r.options.CustomPassword
	conf.Addr = r.options.CustomAddress
	conf.DB = r.options.CustomDB

	return conf
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
	var keepAlive bool

	keepAlive, err = r.client.Expire(ctx, nodePath, ttl).Result()
	if err != nil {
		return err
	}

	register, ok := r.registers.Get(nodePath)
	if ok && register.hash == hv && keepAlive {
		log.Debugf(r.servCtx, "service %q node %q unchanged, skipping registration", serviceName, node.Id)
		return nil
	}

	serviceNode := &discovery.Service{
		Name:     serviceName,
		Nodes:    []discovery.Node{*node},
		Revision: time.Now().UnixMicro(),
	}
	serviceNodeData := encodeService(serviceNode)

	_, err = r.client.Set(ctx, nodePath, serviceNodeData, ttl).Result()
	if err != nil {
		return err
	}

	register = &_Register{
		hash:     hv,
		ttl:      ttl,
		revision: serviceNode.Revision,
	}

	existed := r.registers.Set(nodePath, register, register.revision, 0)
	if existed != register {
		return nil
	}

	log.Debugf(r.servCtx, "register service %q node %q success", serviceNode.Name, node.Id)
	return nil
}

func (r *_Registry) deregisterNode(ctx context.Context, serviceName string, node *discovery.Node) error {
	nodePath := getNodePath(r.options.KeyPrefix, serviceName, node.Id)

	register, ok := r.registers.Get(nodePath)
	if !ok {
		return nil
	}

	r.registers.Del(nodePath, register.revision+1)

	if _, err := r.client.Del(ctx, nodePath).Result(); err != nil {
		return err
	}

	log.Debugf(r.servCtx, "deregister service %q node %q success", serviceName, node.Id)
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
	service := strings.ReplaceAll(s, ":", "-")
	node := strings.ReplaceAll(id.String(), ":", "-")
	return fmt.Sprintf("%s%s:%s", prefix, service, node)
}

func getServicePath(prefix, s string) string {
	service := strings.ReplaceAll(s, ":", "-")
	return fmt.Sprintf("%s%s:*", prefix, service)
}
