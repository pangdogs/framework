package etcd

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"github.com/golaxy-kit/golaxy/service"
	"github.com/golaxy-kit/plugins/logger"
	"github.com/golaxy-kit/plugins/registry"
	hash "github.com/mitchellh/hashstructure"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
	"path"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	prefix = "/golaxy/registry/"
)

func newRegistry(options ...EtcdOption) registry.Registry {
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
	options    EtcdOptions
	serviceCtx service.Context
	client     *clientv3.Client
	register   map[string]uint64
	leases     map[string]clientv3.LeaseID
	sync.RWMutex
}

func (r *_EtcdRegistry) Init(ctx service.Context) {
	r.serviceCtx = ctx

	client, err := clientv3.New(r.configure())
	if err != nil {
		panic(err)
	}
	r.client = client
}

func (r *_EtcdRegistry) Shut() {
	if r.client != nil {
		r.client.Close()
	}
}

func (r *_EtcdRegistry) Register(ctx context.Context, service registry.Service, options ...registry.RegisterOption) error {
	if len(service.Nodes) <= 0 {
		return errors.New("require at least one node")
	}

	var opts registry.RegisterOptions

	for i := range options {
		options[i](&opts)
	}

	var anyErr error

	for _, node := range service.Nodes {
		if err := r.registerNode(ctx, service, node, opts.TTL); err != nil {
			anyErr = err
		}
	}

	return anyErr
}

func (r *_EtcdRegistry) Deregister(ctx context.Context, service registry.Service) error {
	if len(service.Nodes) <= 0 {
		return errors.New("require at least one node")
	}

	var anyErr error

	for _, node := range service.Nodes {
		np := nodePath(service.Name, node.Id)

		r.Lock()
		// delete our hash of the service
		delete(r.register, np)
		// delete our lease of the service
		delete(r.leases, np)
		r.Unlock()

		ctx, cancel := context.WithTimeout(ctx, r.options.Timeout)
		defer cancel()

		logger.Trace(r.serviceCtx, "deregistering %s id %s", service.Name, node.Id)

		_, err := r.client.Delete(ctx, np)
		if err != nil {
			anyErr = err
		}
	}

	return anyErr
}

func (r *_EtcdRegistry) GetService(ctx context.Context, serviceName string) ([]registry.Service, error) {
	ctx, cancel := context.WithTimeout(ctx, r.options.Timeout)
	defer cancel()

	rsp, err := r.client.Get(ctx, servicePath(serviceName)+"/", clientv3.WithPrefix(), clientv3.WithSerializable())
	if err != nil {
		return nil, err
	}

	if len(rsp.Kvs) <= 0 {
		return nil, registry.ErrNotFound
	}

	serviceMap := map[string]*registry.Service{}

	for _, n := range rsp.Kvs {
		sn := decode(n.Value)
		if sn == nil {
			continue
		}

		s, ok := serviceMap[sn.Version]
		if !ok {
			serviceMap[s.Version] = sn
			continue
		}

		s.Nodes = append(s.Nodes, sn.Nodes...)
	}

	services := make([]registry.Service, 0, len(serviceMap))
	for _, service := range serviceMap {
		services = append(services, *service)
	}

	// sort the services
	sort.Slice(services, func(i, j int) bool { return services[i].Version < services[j].Version })

	return services, nil
}

func (r *_EtcdRegistry) ListServices(ctx context.Context) ([]registry.Service, error) {
	ctx, cancel := context.WithTimeout(ctx, r.options.Timeout)
	defer cancel()

	rsp, err := r.client.Get(ctx, prefix, clientv3.WithPrefix(), clientv3.WithSerializable())
	if err != nil {
		return nil, err
	}

	if len(rsp.Kvs) <= 0 {
		return []registry.Service{}, nil
	}

	versions := make(map[string]*registry.Service)

	for _, n := range rsp.Kvs {
		sn := decode(n.Value)
		if sn == nil {
			continue
		}

		sv := sn.Name + "/" + sn.Version

		v, ok := versions[sv]
		if !ok {
			versions[sv] = sn
			continue
		}

		// append to service:version nodes
		v.Nodes = append(v.Nodes, sn.Nodes...)
	}

	services := make([]registry.Service, 0, len(versions))
	for _, service := range versions {
		services = append(services, *service)
	}

	// sort the services
	sort.Slice(services, func(i, j int) bool { return services[i].Name < services[j].Name })

	return services, nil
}

func (r *_EtcdRegistry) Watch(ctx context.Context, serviceName string) (registry.Watcher, error) {
	return newEtcdWatcher(ctx, r, r.options.Timeout, serviceName)
}

func (r *_EtcdRegistry) configure() clientv3.Config {
	if r.options.EtcdConfig != nil {
		return *r.options.EtcdConfig
	}

	config := clientv3.Config{
		Endpoints:   r.options.Endpoints,
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
	if len(service.Nodes) <= 0 {
		return errors.New("require at least one node")
	}

	np := nodePath(service.Name, node.Id)

	// check existing lease cache
	r.RLock()
	leaseID, ok := r.leases[np]
	r.RUnlock()

	if !ok {
		// missing lease, check if the key exists
		ctx, cancel := context.WithTimeout(ctx, r.options.Timeout)
		defer cancel()

		// look for the existing key
		rsp, err := r.client.Get(ctx, np, clientv3.WithSerializable())
		if err != nil {
			return err
		}

		// get the existing lease
		for _, kv := range rsp.Kvs {
			if kv.Lease > 0 {
				leaseID = clientv3.LeaseID(kv.Lease)

				// decode the existing node
				srv := decode(kv.Value)
				if srv == nil || len(srv.Nodes) <= 0 {
					continue
				}

				// create hash of service; uint64
				h, err := hash.Hash(srv.Nodes[0], nil)
				if err != nil {
					continue
				}

				// save the info
				r.Lock()
				r.leases[np] = leaseID
				r.register[np] = h
				r.Unlock()

				break
			}
		}
	}

	var leaseNotFound bool

	// renew the lease if it exists
	if leaseID > 0 {
		logger.Tracef(r.serviceCtx, "renewing existing lease for %s %d", service.Name, leaseID)

		if _, err := r.client.KeepAliveOnce(context.Background(), leaseID); err != nil {
			if err != rpctypes.ErrLeaseNotFound {
				return err
			}

			logger.Tracef(r.serviceCtx, "lease not found for %s %d", service.Name, leaseID)
			// lease not found do register
			leaseNotFound = true
		}
	}

	// create hash of service; uint64
	h, err := hash.Hash(node, nil)
	if err != nil {
		return err
	}

	// get existing hash for the service node
	r.Lock()
	v, ok := r.register[np]
	r.Unlock()

	// the service is unchanged, skip registering
	if ok && v == h && !leaseNotFound {
		logger.Tracef(r.serviceCtx, "service %s node %s unchanged skipping registration", service.Name, node.Id)
		return nil
	}

	nodeService := &registry.Service{
		Name:      service.Name,
		Version:   service.Version,
		Metadata:  service.Metadata,
		Endpoints: service.Endpoints,
		Nodes:     []registry.Node{node},
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.options.Timeout)
	defer cancel()

	var lgr *clientv3.LeaseGrantResponse
	if ttl.Seconds() > 0 {
		// get a lease used to expire keys since we have a ttl
		lgr, err = r.client.Grant(ctx, int64(ttl.Seconds()))
		if err != nil {
			return err
		}
	}

	logger.Tracef(r.serviceCtx, "registering %s id %s with lease %v and leaseID %v and ttl %v", nodeService.Name, node.Id, lgr, lgr.ID, ttl)
	// create an entry for the node
	if lgr != nil {
		_, err = r.client.Put(ctx, np, encode(nodeService), clientv3.WithLease(lgr.ID))
	} else {
		_, err = r.client.Put(ctx, np, encode(nodeService))
	}
	if err != nil {
		return err
	}

	r.Lock()
	// save our hash of the service
	r.register[np] = h
	// save our leaseID of the service
	if lgr != nil {
		r.leases[np] = lgr.ID
	}
	r.Unlock()

	return nil
}

func encode(s *registry.Service) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func decode(ds []byte) *registry.Service {
	var s *registry.Service
	json.Unmarshal(ds, &s)
	return s
}

func nodePath(s, id string) string {
	service := strings.ReplaceAll(s, "/", "-")
	node := strings.ReplaceAll(id, "/", "-")
	return path.Join(prefix, service, node)
}

func servicePath(s string) string {
	return path.Join(prefix, strings.ReplaceAll(s, "/", "-"))
}
