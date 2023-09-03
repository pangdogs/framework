package redis_registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	hash "github.com/mitchellh/hashstructure/v2"
	"github.com/redis/go-redis/v9"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util"
	"kit.golaxy.org/plugins/logger"
	"kit.golaxy.org/plugins/registry"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
)

// NewRegistry 导出newRedisRegistry，可以配合cache registry将数据缓存本地，提高查询效率
var NewRegistry = newRedisRegistry

func newRedisRegistry(options ...RegistryOption) registry.Registry {
	opts := RegistryOptions{}
	Option{}.Default()(&opts)

	for i := range options {
		options[i](&opts)
	}

	return &_RedisRegistry{
		options:  opts,
		register: map[string]uint64{},
	}
}

type _RedisRegistry struct {
	options  RegistryOptions
	ctx      service.Context
	client   *redis.Client
	register map[string]uint64
	mutex    sync.RWMutex
}

// InitSP 初始化服务插件
func (r *_RedisRegistry) InitSP(ctx service.Context) {
	logger.Infof(ctx, "init service plugin %q with %q", definePlugin.Name, util.TypeOfAnyFullName(*r))

	r.ctx = ctx

	if r.options.RedisClient == nil {
		r.client = redis.NewClient(r.configure())
	} else {
		r.client = r.options.RedisClient
	}

	_, err := r.client.Ping(ctx).Result()
	if err != nil {
		log.Panicf("ping redis %q failed, %v", r.client, err)
	}

	_, err = r.client.ConfigSet(ctx, "notify-keyspace-events", "KEA").Result()
	if err != nil {
		log.Panicf("redis %q enable notify-keyspace-events failed, %v", r.client, err)
	}
}

// ShutSP 关闭服务插件
func (r *_RedisRegistry) ShutSP(ctx service.Context) {
	logger.Infof(ctx, "shut service plugin %q", definePlugin.Name)

	if r.options.RedisClient == nil {
		if r.client != nil {
			r.client.Close()
		}
	}
}

// Register 注册服务
func (r *_RedisRegistry) Register(ctx context.Context, service registry.Service, ttl time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if len(service.Nodes) <= 0 {
		return errors.New("require at least one node")
	}

	var errorList []error

	for _, node := range service.Nodes {
		if err := r.registerNode(ctx, service, node, ttl); err != nil {
			errorList = append(errorList, fmt.Errorf("%s:%s", node.Id, err))
		}
	}

	return errors.Join(errorList...)
}

// Deregister 取消注册服务
func (r *_RedisRegistry) Deregister(ctx context.Context, service registry.Service) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if len(service.Nodes) <= 0 {
		return errors.New("require at least one node")
	}

	var errorList []error

	for _, node := range service.Nodes {
		if err := r.deregisterNode(ctx, service, node); err != nil {
			errorList = append(errorList, fmt.Errorf("%s:%s", node.Id, err))
		}
	}

	return errors.Join(errorList...)
}

// GetServiceNode 查询服务节点
func (r *_RedisRegistry) GetServiceNode(ctx context.Context, serviceName, nodeId string) (*registry.Service, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if serviceName == "" || nodeId == "" {
		return nil, registry.ErrNotFound
	}

	nodeVal, err := r.client.Get(ctx, getNodePath(r.options.KeyPrefix, serviceName, nodeId)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, registry.ErrNotFound
		}
		return nil, err
	}

	return decodeService(nodeVal)
}

// GetService 查询服务
func (r *_RedisRegistry) GetService(ctx context.Context, serviceName string) ([]registry.Service, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if serviceName == "" {
		return nil, registry.ErrNotFound
	}

	nodeKeys, err := r.client.Keys(ctx, getServicePath(r.options.KeyPrefix, serviceName)).Result()
	if err != nil {
		return nil, err
	}

	if len(nodeKeys) <= 0 {
		return nil, registry.ErrNotFound
	}

	nodeVals, err := r.client.MGet(ctx, nodeKeys...).Result()
	if err != nil {
		return nil, err
	}

	serviceMap := map[string]*registry.Service{}

	for _, v := range nodeVals {
		service, err := decodeService([]byte(v.(string)))
		if err != nil {
			logger.Error(r.ctx, err)
			continue
		}

		s, ok := serviceMap[service.Version]
		if !ok {
			serviceMap[s.Version] = service
			continue
		}

		s.Nodes = append(s.Nodes, service.Nodes...)
	}

	services := make([]registry.Service, 0, len(serviceMap))
	for _, service := range serviceMap {
		services = append(services, *service)
	}

	// sort the services
	sort.Slice(services, func(i, j int) bool {
		return services[i].Version < services[j].Version
	})

	return services, nil
}

// ListServices 查询所有服务
func (r *_RedisRegistry) ListServices(ctx context.Context) ([]registry.Service, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	nodeKeys, err := r.client.Keys(ctx, r.options.KeyPrefix+"*").Result()
	if err != nil {
		return nil, err
	}

	if len(nodeKeys) <= 0 {
		return nil, nil
	}

	nodeVals, err := r.client.MGet(ctx, nodeKeys...).Result()
	if err != nil {
		return nil, err
	}

	versions := make(map[string]*registry.Service)

	for _, v := range nodeVals {
		service, err := decodeService([]byte(v.(string)))
		if err != nil {
			logger.Error(r.ctx, err)
			continue
		}

		version := service.Name + ":" + service.Version

		s, ok := versions[version]
		if !ok {
			versions[version] = service
			continue
		}

		// append to service:version nodes
		s.Nodes = append(s.Nodes, service.Nodes...)
	}

	services := make([]registry.Service, 0, len(versions))
	for _, service := range versions {
		services = append(services, *service)
	}

	// sort the services
	sort.Slice(services, func(i, j int) bool {
		if services[i].Name == services[j].Name {
			return services[i].Version < services[j].Version
		}
		return services[i].Name < services[j].Name
	})

	return services, nil
}

// Watch 获取服务监听器
func (r *_RedisRegistry) Watch(ctx context.Context, serviceName string) (registry.Watcher, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	return newRedisWatcher(ctx, r, serviceName)
}

func (r *_RedisRegistry) configure() *redis.Options {
	if r.options.RedisConfig != nil {
		return r.options.RedisConfig
	}

	if r.options.RedisURL != "" {
		conf, err := redis.ParseURL(r.options.RedisURL)
		if err != nil {
			logger.Panicf(r.ctx, "parse redis url %q failed, %s", r.options.RedisURL, err)
		}
		return conf
	}

	conf := &redis.Options{}
	conf.Username = r.options.FastUsername
	conf.Password = r.options.FastPassword
	conf.Addr = r.options.FastAddress
	conf.DB = r.options.FastDBIndex

	return conf
}

func (r *_RedisRegistry) registerNode(ctx context.Context, service registry.Service, node registry.Node, ttl time.Duration) error {
	if service.Name == "" {
		return errors.New("service name can't empty")
	}

	if node.Id == "" {
		return errors.New("service node id can't empty")
	}

	if ttl < 0 {
		ttl = 0
	}

	hv, err := hash.Hash(node, hash.FormatV2, nil)
	if err != nil {
		return err
	}

	nodePath := getNodePath(r.options.KeyPrefix, service.Name, node.Id)
	var keepAlive bool

	if ttl.Seconds() > 0 {
		keepAlive, err = r.client.Expire(ctx, nodePath, ttl).Result()
		if err != nil {
			return err
		}

		logger.Debugf(r.ctx, "renewing existing %q id %q with ttl %q, result %t", service.Name, node.Id, ttl, keepAlive)
	}

	r.mutex.RLock()
	rhv, ok := r.register[nodePath]
	r.mutex.RUnlock()

	if ok && rhv == hv && keepAlive {
		logger.Debugf(r.ctx, "service %q node %q unchanged skipping registration", service.Name, node.Id)
		return nil
	}

	serviceNode := service
	serviceNode.Nodes = []registry.Node{node}
	serviceNodeData := encodeService(&serviceNode)

	logger.Debugf(r.ctx, "registering %q id %q content %q with ttl %q", serviceNode.Name, node.Id, serviceNodeData, ttl)

	_, err = r.client.Set(ctx, nodePath, serviceNodeData, ttl).Result()
	if err != nil {
		return err
	}

	r.mutex.Lock()
	r.register[nodePath] = hv
	r.mutex.Unlock()

	return nil
}

func (r *_RedisRegistry) deregisterNode(ctx context.Context, service registry.Service, node registry.Node) error {
	logger.Debugf(r.ctx, "deregistering %q id %q", service.Name, node.Id)

	nodePath := getNodePath(r.options.KeyPrefix, service.Name, node.Id)

	r.mutex.Lock()
	delete(r.register, nodePath)
	r.mutex.Unlock()

	_, err := r.client.Del(ctx, nodePath).Result()
	return err
}

func encodeService(s *registry.Service) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func decodeService(ds []byte) (s *registry.Service, err error) {
	return s, json.Unmarshal(ds, &s)
}

func getNodePath(prefix, s, id string) string {
	service := strings.ReplaceAll(s, ":", "-")
	node := strings.ReplaceAll(id, ":", "-")
	return fmt.Sprintf("%s%s:%s", prefix, service, node)
}

func getServicePath(prefix, s string) string {
	service := strings.ReplaceAll(s, ":", "-")
	return fmt.Sprintf("%s%s:*", prefix, service)
}