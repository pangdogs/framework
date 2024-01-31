package cache_discovery

import (
	"context"
	"errors"
	"git.golaxy.org/core"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/framework/plugins/discovery"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/plugins/util/concurrent"
	"github.com/elliotchance/pie/v2"
	"sync"
	"time"
)

func newRegistry(settings ...option.Setting[RegistryOptions]) discovery.IRegistry {
	return &_Registry{
		options:    option.Make(Option{}.Default(), settings...),
		serviceMap: concurrent.MakeLockedMap[string, *discovery.Service](0),
	}
}

type _Registry struct {
	discovery.IRegistry
	ctx        context.Context
	cancel     context.CancelFunc
	options    RegistryOptions
	servCtx    service.Context
	wg         sync.WaitGroup
	serviceMap concurrent.LockedMap[string, *discovery.Service]
	revision   int64
}

// InitSP 初始化服务插件
func (r *_Registry) InitSP(ctx service.Context) {
	log.Infof(ctx, "init plugin %q", plugin.Name)

	if r.options.Registry == nil {
		log.Panic(ctx, "wrap registry is nil, must be set before init")
	}
	r.IRegistry = r.options.Registry

	r.ctx, r.cancel = context.WithCancel(context.Background())
	r.servCtx = ctx

	if init, ok := r.IRegistry.(core.LifecycleServicePluginInit); ok {
		init.InitSP(r.servCtx)
	}

	if err := r.refreshCache(); err != nil {
		log.Panicf(r.servCtx, "refresh cache failed, %s", err)
	}

	r.wg.Add(1)
	go r.mainLoop()
}

// ShutSP 关闭服务插件
func (r *_Registry) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut plugin %q", plugin.Name)

	r.cancel()
	r.wg.Wait()

	if shut, ok := r.IRegistry.(core.LifecycleServicePluginShut); ok {
		shut.ShutSP(ctx)
	}
}

// GetServiceNode 查询服务节点
func (r *_Registry) GetServiceNode(ctx context.Context, serviceName, nodeId string) (*discovery.Service, error) {
	services, ok := r.serviceMap.Get(serviceName)
	if !ok {
		return nil, discovery.ErrNotFound
	}

	idx := pie.FindFirstUsing(services.Nodes, func(value discovery.Node) bool {
		return value.Id == nodeId
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
	services, ok := r.serviceMap.Get(serviceName)
	if !ok {
		return nil, discovery.ErrNotFound
	}

	return services, nil
}

// ListServices 查询所有服务
func (r *_Registry) ListServices(ctx context.Context) ([]discovery.Service, error) {
	allServices := make([]discovery.Service, 0, r.serviceMap.Len())

	r.serviceMap.AutoRLock(func(allServicesMap *map[string]*discovery.Service) {
		for _, value := range *allServicesMap {
			allServices = append(allServices, *value)
		}
	})

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
			if errors.Is(err, discovery.ErrStoppedWatching) {
				time.Sleep(retryInterval)
				goto retry
			}

			log.Errorf(r.servCtx, "watching service changes failed, %s, retry it", err)
			<-watcher.Stop()
			time.Sleep(retryInterval)
			goto retry
		}

		r.updateCache(event)
	}

end:
	if watcher != nil {
		<-watcher.Stop()
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
		r.serviceMap.Insert(service.Name, &service)
	}

	return nil
}

func (r *_Registry) updateCache(event *discovery.Event) {
	if r.revision < event.Service.Revision {
		r.revision = event.Service.Revision
	}

	switch event.Type {
	case discovery.Create, discovery.Update:
		service, ok := r.serviceMap.Get(event.Service.Name)
		if !ok {
			r.serviceMap.Insert(event.Service.Name, event.Service)
			return
		}

		serviceCopy := service.DeepCopy()

		idx := pie.FindFirstUsing(serviceCopy.Nodes, func(value discovery.Node) bool {
			return value.Id == event.Service.Nodes[0].Id
		})
		if idx < 0 {
			serviceCopy.Nodes = append(serviceCopy.Nodes, event.Service.Nodes[0])
			r.serviceMap.Insert(serviceCopy.Name, serviceCopy)
			return
		}

		serviceCopy.Nodes[idx] = event.Service.Nodes[0]
		r.serviceMap.Insert(serviceCopy.Name, serviceCopy)
		return

	case discovery.Delete:
		service, ok := r.serviceMap.Get(event.Service.Name)
		if !ok {
			return
		}

		serviceCopy := service.DeepCopy()

		idx := pie.FindFirstUsing(serviceCopy.Nodes, func(value discovery.Node) bool {
			return value.Id == event.Service.Nodes[0].Id
		})
		if idx < 0 {
			return
		}

		pie.Delete(serviceCopy.Nodes, idx)
		r.serviceMap.Insert(serviceCopy.Name, serviceCopy)
		return
	}
}
