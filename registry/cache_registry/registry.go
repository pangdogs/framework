package cache_registry

import (
	"context"
	"errors"
	"kit.golaxy.org/golaxy"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util/option"
	"kit.golaxy.org/golaxy/util/types"
	"kit.golaxy.org/plugins/log"
	"kit.golaxy.org/plugins/registry"
	"reflect"
	"sync"
)

func newRegistry(settings ...option.Setting[RegistryOptions]) registry.Registry {
	return &_Registry{
		options:        option.Make(Option{}.Default(), settings...),
		serviceMap:     map[string]*[]registry.Service{},
		serviceNodeMap: map[[2]string]*registry.Service{},
	}
}

type _Registry struct {
	registry.Registry
	ctx            context.Context
	cancel         context.CancelFunc
	options        RegistryOptions
	servCtx        service.Context
	wg             sync.WaitGroup
	serviceMap     map[string]*[]registry.Service
	serviceNodeMap map[[2]string]*registry.Service
	mutex          sync.RWMutex
}

// InitSP 初始化服务插件
func (r *_Registry) InitSP(ctx service.Context) {
	if r.options.Registry == nil {
		log.Panic(ctx, "wrap registry plugin is nil, must be set before init")
	}
	r.Registry = r.options.Registry

	r.ctx, r.cancel = context.WithCancel(context.Background())
	r.servCtx = ctx

	log.Infof(r.servCtx, "init service plugin <%s>:[%s,%s]", plugin.Name, types.AnyFullName(*r), types.TypeFullName(reflect.TypeOf(r.Registry).Elem()))

	if init, ok := r.Registry.(golaxy.LifecycleServicePluginInit); ok {
		init.InitSP(r.servCtx)
	}

	services, err := r.Registry.ListServices(r.servCtx)
	if err != nil {
		log.Panicf(r.servCtx, "list all services failed, %s", err)
	}

	for i := range services {
		service := &services[i]

		versions := r.getServiceVersions(service.Name)
		*versions = append(*versions, *service)

		for j := range service.Nodes {
			node := &service.Nodes[j]

			serviceNode := *service
			serviceNode.Nodes = []registry.Node{*node}

			r.serviceNodeMap[[2]string{service.Name, node.Id}] = &serviceNode
		}
	}

	watcher, err := r.Registry.Watch(r.ctx, "")
	if err != nil {
		log.Panicf(r.servCtx, "watching service changes failed, %s", err)
	}

	r.wg.Add(1)
	go r.mainLoop(watcher)
}

// ShutSP 关闭服务插件
func (r *_Registry) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut service plugin <%s>:[%s,%s]", plugin.Name, types.AnyFullName(*r), types.TypeFullName(reflect.TypeOf(r.Registry).Elem()))

	r.cancel()
	r.wg.Wait()

	if shut, ok := r.Registry.(golaxy.LifecycleServicePluginShut); ok {
		shut.ShutSP(ctx)
	}
}

// GetServiceNode 查询服务节点
func (r *_Registry) GetServiceNode(ctx context.Context, serviceName, nodeId string) (*registry.Service, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	serviceNode, ok := r.serviceNodeMap[[2]string{serviceName, nodeId}]
	if !ok {
		return nil, registry.ErrNotFound
	}

	return serviceNode.DeepCopy(), nil
}

// GetService 查询服务
func (r *_Registry) GetService(ctx context.Context, serviceName string) ([]registry.Service, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	versions, ok := r.serviceMap[serviceName]
	if !ok {
		return nil, registry.ErrNotFound
	}

	versionsCopy := make([]registry.Service, 0, len(*versions))

	for _, service := range *versions {
		versionsCopy = append(versionsCopy, *service.DeepCopy())
	}

	return versionsCopy, nil
}

// ListServices 查询所有服务
func (r *_Registry) ListServices(ctx context.Context) ([]registry.Service, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	servicesCopy := make([]registry.Service, 0, len(r.serviceNodeMap))

	for _, versions := range r.serviceMap {
		for _, service := range *versions {
			servicesCopy = append(servicesCopy, *service.DeepCopy())
		}
	}

	return servicesCopy, nil
}

func (r *_Registry) getServiceVersions(serviceName string) *[]registry.Service {
	services, ok := r.serviceMap[serviceName]
	if !ok {
		services = &[]registry.Service{}
		r.serviceMap[serviceName] = services
	}
	return services
}

func (r *_Registry) mainLoop(watcher registry.Watcher) {
	defer r.wg.Done()

	log.Debug(r.servCtx, "watching service changes started")

loop:
	for {
		event, err := watcher.Next()
		if err != nil {
			if errors.Is(err, registry.ErrStoppedWatching) {
				break loop
			}
			log.Errorf(r.servCtx, "watching service changes failed, %s", err)
			continue
		}

		func() {
			r.mutex.Lock()
			defer r.mutex.Unlock()

			eventNode := event.Service.Nodes[0]

			removeNode := func(versions *[]registry.Service, versionIdx int, service *registry.Service) {
				for i := len(service.Nodes) - 1; i >= 0; i-- {
					node := &service.Nodes[i]

					if node.Id == eventNode.Id {
						service.Nodes = append(service.Nodes[:i], service.Nodes[i+1:]...)
					}
				}

				if len(service.Nodes) <= 0 {
					*versions = append((*versions)[:versionIdx], (*versions)[versionIdx+1:]...)
				}
			}

			switch event.Type {
			case registry.Create, registry.Update:
				r.serviceNodeMap[[2]string{event.Service.Name, eventNode.Id}] = event.Service

				versions := r.getServiceVersions(event.Service.Name)

				for i := len(*versions) - 1; i >= 0; i-- {
					service := &(*versions)[i]

					if service.Version == event.Service.Version {
						continue
					}

					removeNode(versions, i, service)
				}

				for i := range *versions {
					service := &(*versions)[i]

					if service.Version != event.Service.Version {
						continue
					}

					for j := range service.Nodes {
						node := &service.Nodes[j]

						if node.Id == eventNode.Id {
							*node = eventNode
							return
						}
					}

					service.Nodes = append(service.Nodes, eventNode)
					return
				}

				*versions = append(*versions, *event.Service)
				return

			case registry.Delete:
				delete(r.serviceNodeMap, [2]string{event.Service.Name, eventNode.Id})

				versions, ok := r.serviceMap[event.Service.Name]
				if !ok {
					return
				}

				for i := len(*versions) - 1; i >= 0; i-- {
					service := &(*versions)[i]

					if service.Version != event.Service.Version {
						continue
					}

					removeNode(versions, i, service)
				}
			}
		}()
	}

	<-watcher.Stop()

	log.Debugf(r.servCtx, "watching service changes stopped")
}
