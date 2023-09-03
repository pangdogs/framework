package cache_registry

import (
	"context"
	"errors"
	"kit.golaxy.org/golaxy"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util"
	"kit.golaxy.org/plugins/logger"
	"kit.golaxy.org/plugins/registry"
	"reflect"
	"sync"
)

func newCacheRegistry(options ...RegistryOption) registry.Registry {
	opts := RegistryOptions{}
	Option{}.Default()(&opts)

	for i := range options {
		options[i](&opts)
	}

	return &_CacheRegistry{
		options:        opts,
		serviceMap:     map[string]*[]registry.Service{},
		serviceNodeMap: map[_ServiceNodeKey]*registry.Service{},
	}
}

type _ServiceNodeKey struct {
	ServiceName, NodeId string
}

type _CacheRegistry struct {
	registry.Registry
	options        RegistryOptions
	serviceMap     map[string]*[]registry.Service
	serviceNodeMap map[_ServiceNodeKey]*registry.Service
	mutex          sync.RWMutex
	wg             sync.WaitGroup
}

// InitSP 初始化服务插件
func (r *_CacheRegistry) InitSP(ctx service.Context) {
	if r.options.Registry == nil {
		logger.Panic(ctx, "cached plugin is nil, must be set before init")
	}
	r.Registry = r.options.Registry

	logger.Infof(ctx, "init service plugin %q with %q, cached %q", definePlugin.Name, util.TypeOfAnyFullName(*r), util.TypeOfFullName(reflect.TypeOf(r.options.Registry).Elem()))

	if init, ok := r.options.Registry.(golaxy.LifecycleServicePluginInit); ok {
		init.InitSP(ctx)
	}

	watcher, err := r.Registry.Watch(ctx, "")
	if err != nil {
		logger.Panicf(ctx, "new service watcher failed, %s", err)
	}

	services, err := r.Registry.ListServices(ctx)
	if err != nil {
		logger.Panicf(ctx, "list all services failed, %s", err)
	}

	for i := range services {
		service := &services[i]

		versions := r.getServiceVersions(service.Name)
		*versions = append(*versions, *service)

		for j := range service.Nodes {
			node := &service.Nodes[j]

			serviceNode := *service
			serviceNode.Nodes = []registry.Node{*node}

			r.serviceNodeMap[_ServiceNodeKey{
				ServiceName: service.Name,
				NodeId:      node.Id,
			}] = &serviceNode
		}
	}

	r.wg.Add(1)

	go func() {
		defer r.wg.Done()

		for {
			event, err := watcher.Next()
			if err != nil {
				if errors.Is(err, registry.ErrStoppedWatching) {
					logger.Debugf(ctx, "watch service changes stopped")
					return
				}
				logger.Errorf(ctx, "an error occurred during watch service changes, %s", err)
				return
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
					r.serviceNodeMap[_ServiceNodeKey{
						ServiceName: event.Service.Name,
						NodeId:      eventNode.Id,
					}] = event.Service

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
					delete(r.serviceNodeMap, _ServiceNodeKey{
						ServiceName: event.Service.Name,
						NodeId:      eventNode.Id,
					})

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
	}()
}

// ShutSP 关闭服务插件
func (r *_CacheRegistry) ShutSP(ctx service.Context) {
	logger.Infof(ctx, "shut service plugin %q", definePlugin.Name)

	r.wg.Wait()

	if shut, ok := r.options.Registry.(golaxy.LifecycleServicePluginShut); ok {
		shut.ShutSP(ctx)
	}
}

// GetServiceNode 查询服务节点
func (r *_CacheRegistry) GetServiceNode(ctx context.Context, serviceName, nodeId string) (*registry.Service, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	serviceNode, ok := r.serviceNodeMap[_ServiceNodeKey{
		ServiceName: serviceName,
		NodeId:      nodeId,
	}]
	if !ok {
		return nil, registry.ErrNotFound
	}

	return serviceNode.DeepCopy(), nil
}

// GetService 查询服务
func (r *_CacheRegistry) GetService(ctx context.Context, serviceName string) ([]registry.Service, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

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
func (r *_CacheRegistry) ListServices(ctx context.Context) ([]registry.Service, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	servicesCopy := make([]registry.Service, 0, len(r.serviceNodeMap))

	for _, versions := range r.serviceMap {
		for _, service := range *versions {
			servicesCopy = append(servicesCopy, *service.DeepCopy())
		}
	}

	return servicesCopy, nil
}

func (r *_CacheRegistry) getServiceVersions(serviceName string) *[]registry.Service {
	services, ok := r.serviceMap[serviceName]
	if !ok {
		services = &[]registry.Service{}
		r.serviceMap[serviceName] = services
	}
	return services
}