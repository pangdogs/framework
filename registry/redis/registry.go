package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	hash "github.com/mitchellh/hashstructure/v2"
	"github.com/redis/go-redis/v9"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/plugins/logger"
	"kit.golaxy.org/plugins/registry"
	"log"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"
)

func NewRedisRegistry(options ...RedisOption) registry.Registry {
	opts := RedisOptions{}
	WithRedisOption{}.Default()(&opts)

	for i := range options {
		options[i](&opts)
	}

	return &_RedisRegistry{
		options:  opts,
		register: map[string]uint64{},
	}
}

type _RedisRegistry struct {
	options  RedisOptions
	ctx      service.Context
	client   *redis.Client
	register map[string]uint64
	mutex    sync.RWMutex
}

// InitService 初始化服务插件
func (r *_RedisRegistry) InitService(ctx service.Context) {
	logger.Infof(ctx, "init service plugin %q with %q", definePlugin.Name, reflect.TypeOf(_RedisRegistry{}))

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

	_, err = r.client.ConfigSet(ctx, "notify-keyspace-events", "E").Result()
	if err != nil {
		log.Panicf("redis %q enable notify-keyspace-events failed, %v", r.client, err)
	}
}

// ShutService 关闭服务插件
func (r *_RedisRegistry) ShutService(ctx service.Context) {
	logger.Infof(ctx, "shut service plugin %q", definePlugin.Name)

	if r.options.RedisClient == nil {
		if r.client != nil {
			r.client.Close()
		}
	}
}

// Register 注册服务
func (r *_RedisRegistry) Register(ctx context.Context, service registry.Service, ttl time.Duration) error {
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

// GetService 查询服务
func (r *_RedisRegistry) GetService(ctx context.Context, serviceName string) ([]registry.Service, error) {
	if serviceName == "" {
		return nil, registry.ErrNotFound
	}

	var nodeKeys []string
	var err error

	r.invokeWithTimeout(ctx, func(ctx context.Context) {
		nodeKeys, err = r.client.Keys(ctx, getServicePath(r.options.KeyPrefix, serviceName)).Result()
	})
	if err != nil {
		return nil, err
	}

	var nodeVals []any

	r.invokeWithTimeout(ctx, func(ctx context.Context) {
		nodeVals, err = r.client.MGet(ctx, nodeKeys...).Result()
	})
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
	var nodeKeys []string
	var err error

	r.invokeWithTimeout(ctx, func(ctx context.Context) {
		nodeKeys, err = r.client.Keys(ctx, r.options.KeyPrefix+"*").Result()
	})
	if err != nil {
		return nil, err
	}

	var nodeVals []any

	r.invokeWithTimeout(ctx, func(ctx context.Context) {
		nodeVals, err = r.client.MGet(ctx, nodeKeys...).Result()
	})
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
	return newRedisWatcher(ctx, r, serviceName)
}

func (r *_RedisRegistry) configure() *redis.Options {
	if r.options.RedisConfig != nil {
		return r.options.RedisConfig
	}

	conf, err := redis.ParseURL(r.options.RedisURL)
	if err != nil {
		logger.Panicf(r.ctx, "parse redis url %q failed, %s", r.options.RedisURL, err)
	}

	return conf
}

func (r *_RedisRegistry) registerNode(ctx context.Context, service registry.Service, node registry.Node, ttl time.Duration) error {
	if ttl < 0 {
		ttl = 0
	}

	nodePath := getNodePath(r.options.KeyPrefix, service.Name, node.Id)

	nodeService := &registry.Service{
		Name:      service.Name,
		Version:   service.Version,
		Metadata:  service.Metadata,
		Endpoints: service.Endpoints,
		Nodes:     []registry.Node{node},
	}

	hv, err := hash.Hash(nodeService, hash.FormatV2, nil)
	if err != nil {
		return err
	}

	var keepAlive bool

	if ttl.Seconds() > 0 {
		r.invokeWithTimeout(ctx, func(ctx context.Context) {
			keepAlive, err = r.client.Expire(ctx, nodePath, ttl).Result()
		})
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

	nodeServiceData := encodeService(nodeService)

	logger.Debugf(r.ctx, "registering %q id %q content %q with ttl %q", nodeService.Name, node.Id, nodeServiceData, ttl)

	r.invokeWithTimeout(ctx, func(ctx context.Context) {
		_, err = r.client.Set(ctx, nodePath, nodeServiceData, ttl).Result()
	})
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

	var err error

	r.invokeWithTimeout(ctx, func(ctx context.Context) {
		_, err = r.client.Del(ctx, nodePath).Result()
	})

	return err
}

func (r *_RedisRegistry) invokeWithTimeout(ctx context.Context, fun func(ctx context.Context)) {
	if fun == nil {
		return
	}

	if ctx == nil {
		ctx = r.ctx
	}

	if r.options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.options.Timeout)
		defer cancel()
	}

	fun(ctx)
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
