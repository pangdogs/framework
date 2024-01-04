package etcd_registry

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	hash "github.com/mitchellh/hashstructure/v2"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	etcd_client "go.etcd.io/etcd/client/v3"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util/option"
	"kit.golaxy.org/golaxy/util/types"
	"kit.golaxy.org/plugins/log"
	"kit.golaxy.org/plugins/registry"
	"path"
	"sort"
	"strings"
	"sync"
	"time"
)

// NewRegistry 创建registry插件，可以配合cache registry将数据缓存本地，提高查询效率
func NewRegistry(settings ...option.Setting[RegistryOptions]) registry.Registry {
	return &_Registry{
		options:  option.Make(Option{}.Default(), settings...),
		register: make(map[string]uint64),
		leases:   make(map[string]etcd_client.LeaseID),
	}
}

type _Registry struct {
	options  RegistryOptions
	servCtx  service.Context
	client   *etcd_client.Client
	register map[string]uint64
	leases   map[string]etcd_client.LeaseID
	mutex    sync.RWMutex
}

// InitSP 初始化服务插件
func (r *_Registry) InitSP(ctx service.Context) {
	log.Infof(ctx, "init service plugin <%s>:[%s]", plugin.Name, types.AnyFullName(*r))

	r.servCtx = ctx

	if r.options.EtcdClient == nil {
		cli, err := etcd_client.New(r.configure())
		if err != nil {
			log.Panicf(r.servCtx, "new etcd client failed, %s", err)
		}
		r.client = cli
	} else {
		r.client = r.options.EtcdClient
	}

	for _, ep := range r.client.Endpoints() {
		if _, err := r.client.Status(r.servCtx, ep); err != nil {
			log.Panicf(r.servCtx, "status etcd %q failed, %s", ep, err)
		}
	}
}

// ShutSP 关闭服务插件
func (r *_Registry) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut service plugin <%s>:[%s]", plugin.Name, types.AnyFullName(*r))

	if r.options.EtcdClient == nil {
		if r.client != nil {
			r.client.Close()
		}
	}
}

// Register 注册服务
func (r *_Registry) Register(ctx context.Context, service registry.Service, ttl time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if len(service.Nodes) <= 0 {
		return fmt.Errorf("%w: require at least one node", registry.ErrRegistry)
	}

	var errs []error

	for _, node := range service.Nodes {
		if err := r.registerNode(ctx, service, node, ttl); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", node.Id, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w: %w", registry.ErrRegistry, errors.Join(errs...))
	}

	return nil
}

// Deregister 取消注册服务
func (r *_Registry) Deregister(ctx context.Context, service registry.Service) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if len(service.Nodes) <= 0 {
		return fmt.Errorf("%w: require at least one node", registry.ErrRegistry)
	}

	var errs []error

	for _, node := range service.Nodes {
		if err := r.deregisterNode(ctx, service, node); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", node.Id, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w: %w", registry.ErrRegistry, errors.Join(errs...))
	}

	return nil
}

// GetServiceNode 查询服务节点
func (r *_Registry) GetServiceNode(ctx context.Context, serviceName, nodeId string) (*registry.Service, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if serviceName == "" || nodeId == "" {
		return nil, registry.ErrNotFound
	}

	rsp, err := r.client.Get(ctx, getNodePath(r.options.KeyPrefix, serviceName, nodeId), etcd_client.WithSerializable())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", registry.ErrRegistry, err)
	}

	if len(rsp.Kvs) <= 0 {
		return nil, registry.ErrNotFound
	}

	return decodeService(rsp.Kvs[0].Value)
}

// GetService 查询服务
func (r *_Registry) GetService(ctx context.Context, serviceName string) ([]registry.Service, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if serviceName == "" {
		return nil, registry.ErrNotFound
	}

	rsp, err := r.client.Get(ctx, getServicePath(r.options.KeyPrefix, serviceName), etcd_client.WithPrefix(), etcd_client.WithSerializable())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", registry.ErrRegistry, err)
	}

	if len(rsp.Kvs) <= 0 {
		return nil, registry.ErrNotFound
	}

	serviceMap := map[string]*registry.Service{}

	for _, kv := range rsp.Kvs {
		service, err := decodeService(kv.Value)
		if err != nil {
			log.Errorf(r.servCtx, "decode service %q failed, %s", kv, err)
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
func (r *_Registry) ListServices(ctx context.Context) ([]registry.Service, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	rsp, err := r.client.Get(ctx, r.options.KeyPrefix, etcd_client.WithPrefix(), etcd_client.WithSerializable())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", registry.ErrRegistry, err)
	}

	if len(rsp.Kvs) <= 0 {
		return nil, nil
	}

	versions := make(map[string]*registry.Service)

	for _, kv := range rsp.Kvs {
		service, err := decodeService(kv.Value)
		if err != nil {
			log.Errorf(r.servCtx, "decode service %q failed, ", kv.Value, err)
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
func (r *_Registry) Watch(ctx context.Context, pattern string) (registry.Watcher, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	return r.newWatcher(ctx, pattern)
}

func (r *_Registry) configure() etcd_client.Config {
	if r.options.EtcdConfig != nil {
		return *r.options.EtcdConfig
	}

	config := etcd_client.Config{
		Endpoints: r.options.FastAddresses,
		Username:  r.options.FastUsername,
		Password:  r.options.FastPassword,
	}

	if r.options.FastSecure || r.options.FastTLSConfig != nil {
		tlsConfig := r.options.FastTLSConfig
		if tlsConfig == nil {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		config.TLS = tlsConfig
	}

	return config
}

func (r *_Registry) registerNode(ctx context.Context, service registry.Service, node registry.Node, ttl time.Duration) error {
	if service.Name == "" {
		return fmt.Errorf("%w: service name can't empty", registry.ErrRegistry)
	}

	if node.Id == "" {
		return fmt.Errorf("%w: service node id can't empty", registry.ErrRegistry)
	}

	if ttl < 0 {
		ttl = 0
	}

	// create hash of service; uint64
	hv, err := hash.Hash(node, hash.FormatV2, nil)
	if err != nil {
		return fmt.Errorf("%w: %w", registry.ErrRegistry, err)
	}

	nodePath := getNodePath(r.options.KeyPrefix, service.Name, node.Id)

	// check existing lease cache
	r.mutex.RLock()
	leaseID, ok := r.leases[nodePath]
	r.mutex.RUnlock()

	// missing lease, check if the key exists
	if !ok {
		// look for the existing key
		rsp, err := r.client.Get(ctx, nodePath, etcd_client.WithSerializable())
		if err != nil {
			return fmt.Errorf("%w: %w", registry.ErrRegistry, err)
		}

		// get the existing lease
		for _, kv := range rsp.Kvs {
			kvLeaseID := etcd_client.LeaseID(kv.Lease)

			if kvLeaseID != etcd_client.NoLease {
				// decode the existing node
				srv, err := decodeService(kv.Value)
				if err != nil {
					log.Errorf(r.servCtx, "decode service %q failed, %s", kv.Value, err)
					continue
				}

				if len(srv.Nodes) <= 0 {
					log.Errorf(r.servCtx, "decode service %q failed, empty nodes", kv.Value)
					continue
				}

				// create hash of service; uint64
				hv, err := hash.Hash(srv.Nodes[0], hash.FormatV2, nil)
				if err != nil {
					log.Errorf(r.servCtx, "decode service %q failed, %s", kv.Value, err)
					continue
				}

				leaseID = kvLeaseID

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
	if leaseID != etcd_client.NoLease {
		log.Debugf(r.servCtx, "renewing existing lease %d for %q", leaseID, service.Name)

		_, err = r.client.KeepAliveOnce(ctx, leaseID)
		if err != nil {
			if !errors.Is(err, rpctypes.ErrLeaseNotFound) {
				return fmt.Errorf("%w: %w", registry.ErrRegistry, err)
			}
			log.Debugf(r.servCtx, "lease %d not found for %q", leaseID, service.Name)
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
		log.Debugf(r.servCtx, "service %q node %q unchanged skipping registration", service.Name, node.Id)
		return nil
	}

	var lgr *etcd_client.LeaseGrantResponse
	if ttl.Seconds() > 0 {
		// get a lease used to expire keys since we have a ttl
		lgr, err = r.client.Grant(ctx, int64(ttl.Seconds()))
		if err != nil {
			return fmt.Errorf("%w: %w", registry.ErrRegistry, err)
		}
	}

	serviceNode := service
	serviceNode.Nodes = []registry.Node{node}
	serviceNodeData := encodeService(&serviceNode)

	// create an entry for the node
	if lgr != nil {
		log.Debugf(r.servCtx, "registering service %q node %q content %q with lease %q and leaseID %d and ttl %q", serviceNode.Name, node.Id, serviceNodeData, lgr, lgr.ID, ttl)
		_, err = r.client.Put(ctx, nodePath, serviceNodeData, etcd_client.WithLease(lgr.ID))
	} else {
		log.Debugf(r.servCtx, "registering service %q node %q content %q", serviceNode.Name, node.Id, serviceNodeData)
		_, err = r.client.Put(ctx, nodePath, serviceNodeData)
	}
	if err != nil {
		return fmt.Errorf("%w: %w", registry.ErrRegistry, err)
	}

	r.mutex.Lock()
	// save our hash of the service
	r.register[nodePath] = hv
	// save our leaseID of the service
	if lgr != nil {
		r.leases[nodePath] = lgr.ID
	}
	r.mutex.Unlock()

	log.Debugf(r.servCtx, "register service %q node %q success", serviceNode.Name, node.Id)

	return nil
}

func (r *_Registry) deregisterNode(ctx context.Context, service registry.Service, node registry.Node) error {
	log.Debugf(r.servCtx, "deregistering service %q node %q", service.Name, node.Id)

	nodePath := getNodePath(r.options.KeyPrefix, service.Name, node.Id)

	r.mutex.Lock()
	// delete our hash of the service
	delete(r.register, nodePath)
	// delete our lease of the service
	delete(r.leases, nodePath)
	r.mutex.Unlock()

	if _, err := r.client.Delete(ctx, nodePath); err != nil {
		return fmt.Errorf("%w: %w", registry.ErrRegistry, err)
	}

	log.Debugf(r.servCtx, "deregister service %q node %q success", service.Name, node.Id)

	return nil
}

func encodeService(s *registry.Service) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func decodeService(ds []byte) (*registry.Service, error) {
	var s *registry.Service

	if err := json.Unmarshal(ds, &s); err != nil {
		return nil, fmt.Errorf("%w: %w", registry.ErrRegistry, err)
	}

	return s, nil
}

func getNodePath(prefix, s, id string) string {
	service := strings.ReplaceAll(s, "/", "-")
	node := strings.ReplaceAll(id, "/", "-")
	return path.Join(prefix, service, node)
}

func getServicePath(prefix, s string) string {
	service := strings.ReplaceAll(s, "/", "-")
	return path.Join(prefix, service) + "/"
}
