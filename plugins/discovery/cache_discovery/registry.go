package cache_discovery

import (
	"context"
	"errors"
	"git.golaxy.org/core"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/plugins/discovery"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/util/concurrent"
	"slices"
	"sync"
	"time"
)

// newRegistry 创建registry cache插件，在本地缓存其他registry插件返回的数据
func newRegistry(settings ...option.Setting[RegistryOptions]) discovery.IRegistry {
	return &_Registry{
		options: option.Make(With.Default(), settings...),
	}
}

type _Registry struct {
	discovery.IRegistry
	ctx       context.Context
	terminate context.CancelFunc
	options   RegistryOptions
	servCtx   service.Context
	wg        sync.WaitGroup
	cache     *concurrent.Cache[string, *discovery.Service]
	revision  int64
}

// InitSP 初始化服务插件
func (r *_Registry) InitSP(ctx service.Context) {
	log.Infof(ctx, "init self %q", self.Name)

	if r.options.Registry == nil {
		log.Panic(ctx, "wrap registry is nil, must be set before init")
	}
	r.IRegistry = r.options.Registry

	r.ctx, r.terminate = context.WithCancel(context.Background())
	r.servCtx = ctx

	if init, ok := r.IRegistry.(core.LifecycleServicePluginInit); ok {
		init.InitSP(r.servCtx)
	}

	if err := r.refreshCache(); err != nil {
		log.Panicf(r.servCtx, "refresh cache failed, %s", err)
	}

	r.cache = concurrent.NewCache[string, *discovery.Service]()

	r.wg.Add(1)
	go r.mainLoop()
}

// ShutSP 关闭服务插件
func (r *_Registry) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut self %q", self.Name)

	r.terminate()
	r.wg.Wait()

	if shut, ok := r.IRegistry.(core.LifecycleServicePluginShut); ok {
		shut.ShutSP(ctx)
	}
}

// GetServiceNode 查询服务节点
func (r *_Registry) GetServiceNode(ctx context.Context, serviceName string, nodeId uid.Id) (*discovery.Service, error) {
	services, ok := r.cache.Get(serviceName)
	if !ok {
		return nil, discovery.ErrNotFound
	}

	idx := slices.IndexFunc(services.Nodes, func(node discovery.Node) bool {
		return node.Id == nodeId
	})
	if idx < 0 {
		return nil, discovery.ErrNotFound
	}

	service := &discovery.Service{
		Name:     services.Name,
		Nodes:    []discovery.Node{services.Nodes[idx]},
		Revision: services.Revision,
	}

	return service, nil
}

// GetService 查询服务
func (r *_Registry) GetService(ctx context.Context, serviceName string) (*discovery.Service, error) {
	services, ok := r.cache.Get(serviceName)
	if !ok {
		return nil, discovery.ErrNotFound
	}

	return services, nil
}

// ListServices 查询所有服务
func (r *_Registry) ListServices(ctx context.Context) ([]discovery.Service, error) {
	snapshot := r.cache.Snapshot()
	allServices := make([]discovery.Service, 0, len(snapshot))

	for _, kv := range snapshot {
		allServices = append(allServices, *kv.V)
	}

	return allServices, nil
}

func (r *_Registry) mainLoop() {
	defer r.wg.Done()

	log.Debug(r.servCtx, "watching service changes started")

retry:
	var watcher discovery.IWatcher
	var err error
	retryInterval := 3 * time.Second

	select {
	case <-r.ctx.Done():
		goto end
	default:
	}

	watcher, err = r.IRegistry.Watch(r.ctx, "", r.revision)
	if err != nil {
		log.Errorf(r.servCtx, "watching service changes failed, %s, retry it", err)
		time.Sleep(retryInterval)
		goto retry
	}

	for {
		event, err := watcher.Next()
		if err != nil {
			if errors.Is(err, discovery.ErrTerminated) {
				time.Sleep(retryInterval)
				goto retry
			}

			log.Errorf(r.servCtx, "watching service changes failed, %s, retry it", err)
			<-watcher.Terminate()
			time.Sleep(retryInterval)
			goto retry
		}

		r.updateCache(event)
	}

end:
	if watcher != nil {
		<-watcher.Terminate()
	}

	log.Debug(r.servCtx, "watching service changes stopped")
}

func (r *_Registry) refreshCache() error {
	services, err := r.IRegistry.ListServices(r.servCtx)
	if err != nil {
		return err
	}

	for _, service := range services {
		if r.revision < service.Revision {
			r.revision = service.Revision
		}
		r.cache.Set(service.Name, &service, service.Revision, 0)
	}

	return nil
}

func (r *_Registry) updateCache(event *discovery.Event) {
	if r.revision < event.Service.Revision {
		r.revision = event.Service.Revision
	}

	switch event.Type {
	case discovery.Create, discovery.Update:
		service, ok := r.cache.Get(event.Service.Name)
		if !ok {
			r.cache.Set(event.Service.Name, event.Service, service.Revision, 0)
			return
		}

		serviceCopy := service.DeepCopy()

		idx := slices.IndexFunc(serviceCopy.Nodes, func(node discovery.Node) bool {
			return node.Id == event.Service.Nodes[0].Id
		})
		if idx < 0 {
			serviceCopy.Nodes = append(serviceCopy.Nodes, event.Service.Nodes[0])
			r.cache.Set(serviceCopy.Name, serviceCopy, serviceCopy.Revision, 0)
			return
		}

		serviceCopy.Nodes[idx] = event.Service.Nodes[0]
		r.cache.Set(serviceCopy.Name, serviceCopy, serviceCopy.Revision, 0)
		return

	case discovery.Delete:
		service, ok := r.cache.Get(event.Service.Name)
		if !ok {
			return
		}

		serviceCopy := service.DeepCopy()

		idx := slices.IndexFunc(serviceCopy.Nodes, func(node discovery.Node) bool {
			return node.Id == event.Service.Nodes[0].Id
		})
		if idx < 0 {
			return
		}

		serviceCopy.Nodes = slices.Delete(serviceCopy.Nodes, idx, idx+1)
		r.cache.Set(serviceCopy.Name, serviceCopy, serviceCopy.Revision, 0)
		return
	}
}
