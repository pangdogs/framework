/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package cache_discovery

import (
	"context"
	"errors"
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/addins/discovery"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/utils/concurrent"
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
	svcCtx    service.Context
	ctx       context.Context
	terminate context.CancelFunc
	options   RegistryOptions
	wg        sync.WaitGroup
	cache     *concurrent.Cache[string, *discovery.Service]
	revision  int64
}

// Init 初始化插件
func (r *_Registry) Init(svcCtx service.Context, _ runtime.Context) {
	log.Infof(svcCtx, "init addin %q", self.Name)

	if r.options.Registry == nil {
		log.Panic(svcCtx, "wrap registry is nil, must be set before init")
	}
	r.IRegistry = r.options.Registry

	r.svcCtx = svcCtx
	r.ctx, r.terminate = context.WithCancel(context.Background())

	if cb, ok := r.IRegistry.(core.LifecycleAddInInit); ok {
		cb.Init(r.svcCtx, nil)
	}

	if err := r.refreshCache(); err != nil {
		log.Panicf(r.svcCtx, "refresh cache failed, %s", err)
	}

	r.cache = concurrent.NewCache[string, *discovery.Service]()

	r.wg.Add(1)
	go r.mainLoop()
}

// Shut 关闭插件
func (r *_Registry) Shut(svcCtx service.Context, _ runtime.Context) {
	log.Infof(svcCtx, "shut addin %q", self.Name)

	r.terminate()
	r.wg.Wait()

	if cb, ok := r.IRegistry.(core.LifecycleAddInShut); ok {
		cb.Shut(svcCtx, nil)
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

	log.Debug(r.svcCtx, "watching service changes started")

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
		log.Errorf(r.svcCtx, "watching service changes failed, %s, retry it", err)
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

			log.Errorf(r.svcCtx, "watching service changes failed, %s, retry it", err)
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

	log.Debug(r.svcCtx, "watching service changes stopped")
}

func (r *_Registry) refreshCache() error {
	services, err := r.IRegistry.ListServices(r.svcCtx)
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
			r.cache.Set(event.Service.Name, event.Service, event.Service.Revision, 0)
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
