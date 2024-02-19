package etcd_discovery

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/framework/plugins/discovery"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/util/concurrent"
	"github.com/elliotchance/pie/v2"
	hash "github.com/mitchellh/hashstructure/v2"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	etcd_client "go.etcd.io/etcd/client/v3"
	"math"
	"path"
	"strings"
	"time"
)

// NewRegistry 创建registry插件，可以配合registry cache将数据缓存本地，提高查询效率
func NewRegistry(settings ...option.Setting[RegistryOptions]) discovery.IRegistry {
	return &_Registry{
		options:   option.Make(With.Default(), settings...),
		registers: concurrent.MakeLockedMap[string, _Register](0),
	}
}

type _Register struct {
	Hash    uint64
	LeaseId etcd_client.LeaseID
}

type _Registry struct {
	options   RegistryOptions
	servCtx   service.Context
	client    *etcd_client.Client
	registers concurrent.LockedMap[string, _Register]
}

// InitSP 初始化服务插件
func (r *_Registry) InitSP(ctx service.Context) {
	log.Infof(ctx, "init plugin %q", self.Name)

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
		func() {
			ctx, cancel := context.WithTimeout(r.servCtx, 3*time.Second)
			defer cancel()

			if _, err := r.client.Status(ctx, ep); err != nil {
				log.Panicf(r.servCtx, "status etcd %q failed, %s", ep, err)
			}
		}()
	}
}

// ShutSP 关闭服务插件
func (r *_Registry) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut plugin %q", self.Name)

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
		return fmt.Errorf("%w: %w: serivce is nil", discovery.ErrRegistry, core.ErrArgs)
	}

	if len(service.Nodes) <= 0 {
		return fmt.Errorf("%w: require at least one node", discovery.ErrRegistry)
	}

	var errs []error

	for i := range service.Nodes {
		node := &service.Nodes[i]

		if err := r.registerNode(ctx, service.Name, node, ttl); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", node.Id, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w: %w", discovery.ErrRegistry, errors.Join(errs...))
	}

	return nil
}

// Deregister 取消注册服务
func (r *_Registry) Deregister(ctx context.Context, service *discovery.Service) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if service == nil {
		return fmt.Errorf("%w: %w: serivce is nil", discovery.ErrRegistry, core.ErrArgs)
	}

	if len(service.Nodes) <= 0 {
		return fmt.Errorf("%w: require at least one node", discovery.ErrRegistry)
	}

	var errs []error

	for i := range service.Nodes {
		node := &service.Nodes[i]

		if err := r.deregisterNode(ctx, service.Name, node); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", node.Id, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w: %w", discovery.ErrRegistry, errors.Join(errs...))
	}

	return nil
}

// GetServiceNode 查询服务节点
func (r *_Registry) GetServiceNode(ctx context.Context, serviceName, nodeId string) (*discovery.Service, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if serviceName == "" || nodeId == "" {
		return nil, discovery.ErrNotFound
	}

	rsp, err := r.client.Get(ctx, getNodePath(r.options.KeyPrefix, serviceName, nodeId), etcd_client.WithSerializable())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", discovery.ErrRegistry, err)
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
		etcd_client.WithPrefix(),
		etcd_client.WithSort(etcd_client.SortByModRevision, etcd_client.SortDescend),
		etcd_client.WithSerializable(),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", discovery.ErrRegistry, err)
	}

	if len(rsp.Kvs) <= 0 {
		return nil, discovery.ErrNotFound
	}

	service := &discovery.Service{
		Name:  serviceName,
		Nodes: make([]discovery.Node, 0, len(rsp.Kvs)),
	}

	for _, kv := range rsp.Kvs {
		serviceNode, err := decodeService(kv.Value)
		if err != nil {
			log.Errorf(r.servCtx, "decode service %q failed, %s", kv.Value, err)
			continue
		}

		if len(serviceNode.Nodes) <= 0 {
			log.Errorf(r.servCtx, "decode service %q failed, nodes is empty", kv.Value)
			continue
		}

		if service.Revision < kv.ModRevision {
			service.Revision = kv.ModRevision
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
		etcd_client.WithPrefix(),
		etcd_client.WithSort(etcd_client.SortByModRevision, etcd_client.SortDescend),
		etcd_client.WithSerializable())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", discovery.ErrRegistry, err)
	}

	if len(rsp.Kvs) <= 0 {
		return nil, nil
	}

	var services []discovery.Service

	for _, kv := range rsp.Kvs {
		serviceNode, err := decodeService(kv.Value)
		if err != nil {
			log.Errorf(r.servCtx, "decode service %q failed, %s", kv.Value, err)
			continue
		}

		if len(serviceNode.Nodes) <= 0 {
			log.Errorf(r.servCtx, "decode service %q failed, nodes is empty", kv.Value)
			continue
		}

		serviceNode.Revision = kv.ModRevision

		idx := pie.FindFirstUsing(services, func(value discovery.Service) bool {
			return value.Name == serviceNode.Name
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

// Watch 获取服务监听器
func (r *_Registry) Watch(ctx context.Context, pattern string, revision ...int64) (discovery.IWatcher, error) {
	return r.newWatcher(ctx, pattern, revision...)
}

func (r *_Registry) configure() etcd_client.Config {
	if r.options.EtcdConfig != nil {
		return *r.options.EtcdConfig
	}

	config := etcd_client.Config{
		Endpoints:   r.options.CustomAddresses,
		Username:    r.options.CustomUsername,
		Password:    r.options.CustomPassword,
		DialTimeout: 3 * time.Second,
	}

	if r.options.CustomSecure || r.options.CustomTLSConfig != nil {
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

	if ttl < 0 {
		ttl = 0
	}

	// create hash of service; uint64
	hv, err := hash.Hash(node, hash.FormatV2, nil)
	if err != nil {
		return err
	}

	nodePath := getNodePath(r.options.KeyPrefix, serviceName, node.Id)

	// check existing lease cache
	register, ok := r.registers.Get(nodePath)
	if !ok {
		// look for the existing key
		rsp, err := r.client.Get(ctx, nodePath, etcd_client.WithSerializable())
		if err != nil {
			return err
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
					log.Errorf(r.servCtx, "decode service %q failed, nodes is empty", kv.Value)
					continue
				}

				// create hash of service; uint64
				hv, err := hash.Hash(srv.Nodes[0], hash.FormatV2, nil)
				if err != nil {
					log.Errorf(r.servCtx, "decode service %q failed, %s", kv.Value, err)
					continue
				}

				register = _Register{
					Hash:    hv,
					LeaseId: kvLeaseID,
				}

				// save the info
				r.registers.Insert(nodePath, register)
				break
			}
		}
	}

	var leaseNotFound bool

	// renew the lease if it exists
	if register.LeaseId != etcd_client.NoLease {
		log.Debugf(r.servCtx, "renewing existing lease %d for %q", register.LeaseId, serviceName)

		_, err = r.client.KeepAliveOnce(ctx, register.LeaseId)
		if err != nil {
			if !errors.Is(err, rpctypes.ErrLeaseNotFound) {
				return err
			}
			log.Debugf(r.servCtx, "lease %d not found for %q", register.LeaseId, serviceName)
			// lease not found do registers
			leaseNotFound = true
		}
	}

	// get existing hash for the service node, if the service is unchanged, skip registering
	register, ok = r.registers.Get(nodePath)
	if ok && register.Hash == hv && !leaseNotFound {
		log.Debugf(r.servCtx, "service %q node %q unchanged skipping registration", serviceName, node.Id)
		return nil
	}

	serviceNode := &discovery.Service{
		Name:  serviceName,
		Nodes: []discovery.Node{*node},
	}
	serviceNodeData := encodeService(serviceNode)

	// create an entry for the node
	if ttl.Seconds() > 0 {
		// get a lease used to expire keys since we have a ttl
		lgr, err := r.client.Grant(ctx, int64(math.Ceil(ttl.Seconds())))
		if err != nil {
			return err
		}

		log.Debugf(r.servCtx, "registering service %q node %q content %q with lease %d", serviceNode.Name, node.Id, serviceNodeData, lgr.ID)

		_, err = r.client.Put(ctx, nodePath, serviceNodeData, etcd_client.WithLease(lgr.ID))
		if err != nil {
			return err
		}

		// save register info
		register.Hash = hv
		register.LeaseId = lgr.ID

	} else {
		log.Debugf(r.servCtx, "registering service %q node %q content %q no lease", serviceNode.Name, node.Id, serviceNodeData)

		_, err = r.client.Put(ctx, nodePath, serviceNodeData)
		if err != nil {
			return err
		}

		// save register info
		register.Hash = hv
		register.LeaseId = etcd_client.NoLease
	}

	// save our register of the service
	r.registers.Insert(nodePath, register)

	log.Debugf(r.servCtx, "register service %q node %q success", serviceNode.Name, node.Id)

	return nil
}

func (r *_Registry) deregisterNode(ctx context.Context, serviceName string, node *discovery.Node) error {
	log.Debugf(r.servCtx, "deregistering service %q node %q", serviceName, node.Id)

	nodePath := getNodePath(r.options.KeyPrefix, serviceName, node.Id)

	// delete our register of the service
	r.registers.Delete(nodePath)

	if _, err := r.client.Delete(ctx, nodePath); err != nil {
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
		return nil, fmt.Errorf("%w: %w", discovery.ErrRegistry, err)
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
