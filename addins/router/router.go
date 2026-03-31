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

package router

import (
	"context"
	"crypto/tls"
	"errors"
	"path"
	"sync"
	"sync/atomic"
	"time"

	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/addins/gate"
	"git.golaxy.org/framework/addins/log"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

var (
	ErrEntityNotFound  = errors.New("router: entity not found")
	ErrSessionNotFound = errors.New("router: session not found")
	ErrGroupNotFound   = errors.New("router: group not found")
	ErrGroupExists     = errors.New("router: group already exists")
)

// IRouter 路由器接口
type IRouter interface {
	// Map 添加实体与会话的路由映射
	Map(entityId, sessionId uid.Id) (IMapping, error)
	// Lookup 查询映射，可传实体id或会话id
	Lookup(id uid.Id) (IMapping, bool)
	// AddGroup 创建路由组
	AddGroup(ctx context.Context, name string, ids []uid.Id, ttl time.Duration) (IGroup, error)
	// DeleteGroup 删除路由组
	DeleteGroup(ctx context.Context, name string)
	// GetGroupByName 使用名称查询路由组
	GetGroupByName(ctx context.Context, name string) (IGroup, bool)
	// GetGroupByAddr 使用地址查询路由组
	GetGroupByAddr(ctx context.Context, addr string) (IGroup, bool)
}

func newRouter(settings ...option.Setting[RouterOptions]) IRouter {
	return &_Router{
		options:  option.New(With.Default(), settings...),
		mappings: map[uid.Id]*_Mapping{},
	}
}

type _Router struct {
	svcCtx                 service.Context
	ctx                    context.Context
	terminate              context.CancelFunc
	barrier                generic.Barrier
	options                RouterOptions
	groupIdKeyPrefix       string
	groupEntitiesKeyPrefix string
	entityGroupsKeyPrefix  string
	gate                   gate.IGate
	client                 *etcdv3.Client
	mappingMu              sync.RWMutex
	mappings               map[uid.Id]*_Mapping
	groups                 sync.Map // map[string]*_Group, key is group name
	groupCount             atomic.Int64
}

// Init 初始化插件
func (r *_Router) Init(svcCtx service.Context) {
	log.L(svcCtx).Info("initializing add-in", zap.String("name", AddIn.Name))

	r.svcCtx = svcCtx
	r.ctx, r.terminate = context.WithCancel(context.Background())

	r.groupIdKeyPrefix = path.Join(r.options.GroupKeyPrefix, "id") + "/"
	r.groupEntitiesKeyPrefix = path.Join(r.options.GroupKeyPrefix, "entities") + "/"
	r.entityGroupsKeyPrefix = path.Join(r.options.EntityKeyPrefix, "groups") + "/"

	r.gate = gate.AddIn.Require(r.svcCtx)

	if r.options.EtcdClient == nil {
		cli, err := etcdv3.New(r.configure())
		if err != nil {
			log.L(svcCtx).Panic("new etcd client failed", log.JSON("config", r.configure()), zap.Error(err))
		}
		r.client = cli
	} else {
		r.client = r.options.EtcdClient
	}

	for _, ep := range r.client.Endpoints() {
		func(endpoint string) {
			ctx, cancel := context.WithTimeout(r.svcCtx, 3*time.Second)
			defer cancel()

			if _, err := r.client.Status(ctx, endpoint); err != nil {
				log.L(svcCtx).Panic("status etcd failed", zap.String("endpoint", endpoint), zap.Error(err))
			}
		}(ep)
	}
}

// Shut 关闭插件
func (r *_Router) Shut(svcCtx service.Context) {
	log.L(svcCtx).Info("shutting down add-in", zap.String("name", AddIn.Name))

	r.terminate()

	r.barrier.Close()
	r.barrier.Wait()

	if r.options.EtcdClient == nil && r.client != nil {
		r.client.Close()
	}
}

func (r *_Router) configure() etcdv3.Config {
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
