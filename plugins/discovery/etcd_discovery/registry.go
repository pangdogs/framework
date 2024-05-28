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
	"git.golaxy.org/framework/plugins/discovery"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/util/concurrent"
	"github.com/elliotchance/pie/v2"
	hash "github.com/mitchellh/hashstructure/v2"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"maps"
	"math"
	"path"
	"strings"
	"time"
)

// NewRegistry 创建registry插件，可以配合registry cache将数据缓存本地，提高查询效率
func NewRegistry(settings ...option.Setting[RegistryOptions]) discovery.IRegistry {
	return &_Registry{
		options:   option.Make(With.Default(), settings...),
		registers: concurrent.MakeLockedMap[string, *_Register](0),
	}
}

type _Register struct {
	ctx       context.Context
	terminate context.CancelFunc
	hash      uint64
	leaseId   etcdv3.LeaseID
	revision  int64
}

type _Registry struct {
	options   RegistryOptions
	servCtx   service.Context
	client    *etcdv3.Client
	registers concurrent.LockedMap[string, *_Register]
}

// InitSP 初始化服务插件
func (r *_Registry) InitSP(ctx service.Context) {
	log.Infof(ctx, "init plugin %q", self.Name)

	r.servCtx = ctx

	if r.options.EtcdClient == nil {
		cli, err := etcdv3.New(r.configure())
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

	var copied map[string]*_Register

	r.registers.AutoRLock(func(registers *map[string]*_Register) {
		copied = maps.Clone(*registers)
	})

	var errs []error

	for nodePath, register := range copied {
		if register.leaseId == etcdv3.NoLease {
			continue
		}
		_, err := r.client.KeepAliveOnce(ctx, register.leaseId)
		if err != nil {
			errs = append(errs, fmt.Errorf("keeplive %q failed, %w", nodePath, err))
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

	rsp, err := r.client.Get(ctx, getNodePath(r.options.KeyPrefix, serviceName, nodeId), etcdv3.WithSerializable())
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
		etcdv3.WithSort(etcdv3.SortByModRevision, etcdv3.SortDescend),
		etcdv3.WithSerializable())
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
			log.Errorf(r.servCtx, "decode service %q failed, %s", kv.Value, err)
			continue
		}

		if len(serviceNode.Nodes) <= 0 {
			log.Errorf(r.servCtx, "decode service %q failed, nodes is empty", kv.Value)
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
		etcdv3.WithSort(etcdv3.SortByModRevision, etcdv3.SortDescend),
		etcdv3.WithSerializable())
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

	if ttl <= 0 {
		ttl = r.options.TTL
	}

	hv, err := hash.Hash(node, hash.FormatV2, nil)
	if err != nil {
		return err
	}

	nodePath := getNodePath(r.options.KeyPrefix, serviceName, node.Id)
	var leaseId etcdv3.LeaseID

	register, ok := r.registers.Get(nodePath)
	if ok {
		_, err = r.client.KeepAliveOnce(ctx, register.leaseId)
		if !errors.Is(err, rpctypes.ErrLeaseNotFound) {
			return err
		}
		if err == nil {
			if register.hash == hv {
				log.Debugf(r.servCtx, "service %q node %q unchanged, keep alive lease", serviceName, node.Id)
				return nil
			}
			leaseId = register.leaseId
		}
	}

	if leaseId == etcdv3.NoLease {
		lgr, err := r.client.Grant(ctx, int64(math.Ceil(ttl.Seconds())))
		if err != nil {
			return err
		}
		leaseId = lgr.ID
	}

	servNode := &discovery.Service{
		Name:  serviceName,
		Nodes: []discovery.Node{*node},
	}
	servNodeData := encodeService(servNode)

	rsp, err := r.client.Put(ctx, nodePath, servNodeData, etcdv3.WithLease(leaseId))
	if err != nil {
		return err
	}

	var newer *_Register

	r.registers.AutoLock(func(registers *map[string]*_Register) {
		register, ok := (*registers)[nodePath]
		if ok {
			if register.revision > rsp.Header.Revision {
				return
			}
		}

		ctx, cancel := context.WithCancel(r.servCtx)

		newer = &_Register{
			ctx:       ctx,
			terminate: cancel,
			hash:      hv,
			leaseId:   leaseId,
			revision:  rsp.Header.Revision,
		}
		(*registers)[nodePath] = newer
	})

	if newer != nil && r.options.AutoRefreshTTL {
		rspChan, err := r.client.KeepAlive(newer.ctx, leaseId)
		if err != nil {
			return err
		}
		go func() {
			for range rspChan {
				log.Debugf(r.servCtx, "refresh service %q node %q ttl success", servNode.Name, node.Id)
			}
		}()
	}

	log.Debugf(r.servCtx, "register service %q node %q success", servNode.Name, node.Id)
	return nil
}

func (r *_Registry) deregisterNode(ctx context.Context, serviceName string, node *discovery.Node) error {
	nodePath := getNodePath(r.options.KeyPrefix, serviceName, node.Id)
	var old *_Register

	r.registers.AutoLock(func(registers *map[string]*_Register) {
		register, ok := (*registers)[nodePath]
		if ok {
			old = register

			(*registers)[nodePath] = &_Register{
				hash:     0,
				leaseId:  etcdv3.NoLease,
				revision: register.revision,
			}
			register.terminate()
		}
	})

	if old != nil && old.leaseId != etcdv3.NoLease {
		_, err := r.client.Revoke(ctx, old.leaseId)
		if err != nil {
			return err
		}
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
	service := strings.ReplaceAll(s, "/", "-")
	node := strings.ReplaceAll(id.String(), "/", "-")
	return path.Join(prefix, service, node)
}

func getServicePath(prefix, s string) string {
	service := strings.ReplaceAll(s, "/", "-")
	return path.Join(prefix, service) + "/"
}
