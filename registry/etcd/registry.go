package etcd

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	hash "github.com/mitchellh/hashstructure/v2"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/plugins/logger"
	"kit.golaxy.org/plugins/registry"
	"path"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"
)

func newEtcdRegistry(options ...EtcdOption) registry.Registry {
	opts := EtcdOptions{}
	WithEtcdOption{}.Default()(&opts)

	for i := range options {
		options[i](&opts)
	}

	return &_EtcdRegistry{
		options:  opts,
		register: make(map[string]uint64),
		leases:   make(map[string]clientv3.LeaseID),
	}
}

type _EtcdRegistry struct {
	options  EtcdOptions
	ctx      service.Context
	client   *clientv3.Client
	register map[string]uint64
	leases   map[string]clientv3.LeaseID
	mutex    sync.RWMutex
}

// InitService 初始化服务插件
func (r *_EtcdRegistry) InitService(ctx service.Context) {
	logger.Infof(ctx, "init service plugin %q with %q", definePlugin.Name, reflect.TypeOf(_EtcdRegistry{}))

	r.ctx = ctx

	if r.options.EtcdClient == nil {
		cli, err := clientv3.New(r.configure())
		if err != nil {
			logger.Panic(ctx, err)
		}
		r.client = cli
	} else {
		r.client = r.options.EtcdClient
	}
}

// ShutService 关闭服务插件
func (r *_EtcdRegistry) ShutService(ctx service.Context) {
	logger.Infof(ctx, "shut service plugin %q", definePlugin.Name)

	if r.options.EtcdClient == nil {
		if r.client != nil {
			r.client.Close()
		}
	}
}

// Register 注册服务
func (r *_EtcdRegistry) Register(ctx context.Context, service registry.Service, ttl time.Duration) error {
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
func (r *_EtcdRegistry) Deregister(ctx context.Context, service registry.Service) error {
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
func (r *_EtcdRegistry) GetService(ctx context.Context, serviceName string) ([]registry.Service, error) {
	var rsp *clientv3.GetResponse
	var err error

	r.invokeWithTimeout(ctx, func(ctx context.Context) {
		rsp, err = r.client.Get(ctx, getServicePath(r.options.KeyPrefix, serviceName), clientv3.WithPrefix(), clientv3.WithSerializable())
	})
	if err != nil {
		return nil, err
	}

	if len(rsp.Kvs) <= 0 {
		return nil, registry.ErrNotFound
	}

	serviceMap := map[string]*registry.Service{}

	for _, kv := range rsp.Kvs {
		service, err := decodeService(kv.Value)
		if err != nil {
			logger.Debug(r.ctx, err)
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
func (r *_EtcdRegistry) ListServices(ctx context.Context) ([]registry.Service, error) {
	var rsp *clientv3.GetResponse
	var err error

	r.invokeWithTimeout(ctx, func(ctx context.Context) {
		rsp, err = r.client.Get(ctx, r.options.KeyPrefix, clientv3.WithPrefix(), clientv3.WithSerializable())
	})
	if err != nil {
		return nil, err
	}

	if len(rsp.Kvs) <= 0 {
		return nil, nil
	}

	versions := make(map[string]*registry.Service)

	for _, kv := range rsp.Kvs {
		service, err := decodeService(kv.Value)
		if err != nil {
			logger.Debug(r.ctx, err)
			continue
		}

		version := service.Name + "/" + service.Version

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
func (r *_EtcdRegistry) Watch(ctx context.Context, serviceName string) (registry.Watcher, error) {
	return newEtcdWatcher(ctx, r, r.options.Timeout, serviceName)
}

func (r *_EtcdRegistry) configure() clientv3.Config {
	if r.options.EtcdConfig != nil {
		return *r.options.EtcdConfig
	}

	config := clientv3.Config{
		Endpoints:   r.options.Addresses,
		DialTimeout: r.options.Timeout,
		Username:    r.options.Username,
		Password:    r.options.Password,
	}

	if r.options.Secure || r.options.TLSConfig != nil {
		tlsConfig := r.options.TLSConfig
		if tlsConfig == nil {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		config.TLS = tlsConfig
	}

	return config
}

func (r *_EtcdRegistry) registerNode(ctx context.Context, service registry.Service, node registry.Node, ttl time.Duration) error {
	nodePath := getNodePath(r.options.KeyPrefix, service.Name, node.Id)

	// create hash of service; uint64
	hv, err := hash.Hash(node, hash.FormatV2, nil)
	if err != nil {
		return err
	}

	// check existing lease cache
	r.mutex.RLock()
	leaseID, ok := r.leases[nodePath]
	r.mutex.RUnlock()

	// missing lease, check if the key exists
	if !ok {
		var rsp *clientv3.GetResponse
		var err error

		r.invokeWithTimeout(ctx, func(ctx context.Context) {
			// look for the existing key
			rsp, err = r.client.Get(ctx, nodePath, clientv3.WithSerializable())
		})
		if err != nil {
			return err
		}

		// get the existing lease
		for _, kv := range rsp.Kvs {
			if kv.Lease > 0 {
				// decode the existing node
				srv, err := decodeService(kv.Value)
				if err != nil {
					logger.Debug(r.ctx, err)
					continue
				}

				if len(srv.Nodes) <= 0 {
					logger.Debug(r.ctx, "empty nodes")
					continue
				}

				// create hash of service; uint64
				hv, err := hash.Hash(srv.Nodes[0], hash.FormatV2, nil)
				if err != nil {
					logger.Debug(r.ctx, err)
					continue
				}

				leaseID = clientv3.LeaseID(kv.Lease)

				// save the info
				r.mutex.Lock()
				r.leases[nodePath] = leaseID
				r.register[nodePath] = hv
				r.mutex.Unlock()

				break
			}
		}
	}

	var leaseNotFound bool

	// renew the lease if it exists
	if leaseID > 0 {
		logger.Debugf(r.ctx, "renewing existing lease %d for %q", leaseID, service.Name)

		var err error

		r.invokeWithTimeout(ctx, func(ctx context.Context) {
			_, err = r.client.KeepAliveOnce(ctx, leaseID)
		})
		if err != nil {
			if err != rpctypes.ErrLeaseNotFound {
				return err
			}

			logger.Debugf(r.ctx, "lease %d not found for %q", leaseID, service.Name)
			// lease not found do register
			leaseNotFound = true
		}
	}

	// get existing hash for the service node
	r.mutex.RLock()
	v, ok := r.register[nodePath]
	r.mutex.RUnlock()

	// the service is unchanged, skip registering
	if ok && v == hv && !leaseNotFound {
		logger.Debugf(r.ctx, "service %q node %q unchanged skipping registration", service.Name, node.Id)
		return nil
	}

	nodeService := &registry.Service{
		Name:      service.Name,
		Version:   service.Version,
		Metadata:  service.Metadata,
		Endpoints: service.Endpoints,
		Nodes:     []registry.Node{node},
	}

	var lgr *clientv3.LeaseGrantResponse
	if ttl.Seconds() > 0 {
		r.invokeWithTimeout(ctx, func(ctx context.Context) {
			// get a lease used to expire keys since we have a ttl
			lgr, err = r.client.Grant(ctx, int64(ttl.Seconds()))
		})
		if err != nil {
			return err
		}
	}

	nodeServiceData := encodeService(nodeService)

	logger.Debugf(r.ctx, "registering %q id %q content %q with lease %q and leaseID %d and ttl %q", nodeService.Name, node.Id, nodeServiceData, lgr, lgr.ID, ttl)

	// create an entry for the node
	r.invokeWithTimeout(ctx, func(ctx context.Context) {
		if lgr != nil {
			_, err = r.client.Put(ctx, nodePath, nodeServiceData, clientv3.WithLease(lgr.ID))
		} else {
			_, err = r.client.Put(ctx, nodePath, nodeServiceData)
		}
	})
	if err != nil {
		return err
	}

	r.mutex.Lock()
	// save our hash of the service
	r.register[nodePath] = hv
	// save our leaseID of the service
	if lgr != nil {
		r.leases[nodePath] = lgr.ID
	}
	r.mutex.Unlock()

	return nil
}

func (r *_EtcdRegistry) deregisterNode(ctx context.Context, service registry.Service, node registry.Node) error {
	logger.Debugf(r.ctx, "deregistering %q id %q", service.Name, node.Id)

	nodePath := getNodePath(r.options.KeyPrefix, service.Name, node.Id)

	r.mutex.Lock()
	// delete our hash of the service
	delete(r.register, nodePath)
	// delete our lease of the service
	delete(r.leases, nodePath)
	r.mutex.Unlock()

	var err error

	r.invokeWithTimeout(ctx, func(ctx context.Context) {
		_, err = r.client.Delete(ctx, nodePath)
	})

	return err
}

func (r *_EtcdRegistry) invokeWithTimeout(ctx context.Context, fun func(ctx context.Context)) {
	if fun == nil {
		return
	}

	if ctx == nil {
		ctx = r.ctx
	}

	ctx, cancel := context.WithTimeout(ctx, r.options.Timeout)
	defer cancel()

	fun(ctx)
}

func encodeService(s *registry.Service) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func decodeService(ds []byte) (*registry.Service, error) {
	var s *registry.Service
	err := json.Unmarshal(ds, &s)
	return s, err
}

func getNodePath(prefix, s, id string) string {
	service := strings.ReplaceAll(s, "/", "-")
	node := strings.ReplaceAll(id, "/", "-")
	return path.Join(prefix, service, node)
}

func getServicePath(prefix, s string) string {
	return path.Join(prefix, strings.ReplaceAll(s, "/", "-")) + "/"
}
