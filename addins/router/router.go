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
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/addins/gate"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/utils/concurrent"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"time"
)

var (
	ErrEntityNotFound  = errors.New("router: entity not found")
	ErrSessionNotFound = errors.New("router: session not found")
	ErrEntityMapped    = errors.New("router: entity is already mapping")
	ErrSessionMapped   = errors.New("router: session is already mapping")
)

// IRouter 路由器接口
type IRouter interface {
	// Mapping 路由映射
	Mapping(entityId, sessionId uid.Id) (IMapping, error)
	// CleanEntity 清理实体路由信息
	CleanEntity(entityId uid.Id)
	// CleanSession 清理会话路由信息
	CleanSession(sessionId uid.Id)
	// LookupEntity 查找实体
	LookupEntity(sessionId uid.Id) (ec.ConcurrentEntity, string, bool)
	// LookupSession 查找会话
	LookupSession(entityId uid.Id) (gate.ISession, bool)
	// AddGroup 添加分组
	AddGroup(ctx context.Context, name string) (IGroup, error)
	// DeleteGroup 删除分组
	DeleteGroup(ctx context.Context, name string)
	// GetGroup 查询分组
	GetGroup(ctx context.Context, name string) (IGroup, bool)
	// GetGroupByAddr 使用分组地址查询分组
	GetGroupByAddr(ctx context.Context, addr string) (IGroup, bool)
	// RangeGroups 遍历包含实体的所有分组
	RangeGroups(ctx context.Context, entityId uid.Id, fun generic.Func1[IGroup, bool])
	// EachGroups 遍历包含实体的所有分组
	EachGroups(ctx context.Context, entityId uid.Id, fun generic.Action1[IGroup])
}

func newRouter(settings ...option.Setting[RouterOptions]) IRouter {
	return &_Router{
		options:  option.Make(With.Default(), settings...),
		planning: concurrent.MakeLockedMap[uid.Id, *_Mapping](0),
	}
}

type _Router struct {
	svcCtx            service.Context
	options           RouterOptions
	gate              gate.IGate
	client            *etcdv3.Client
	planning          concurrent.LockedMap[uid.Id, *_Mapping]
	groupCache        *concurrent.Cache[string, *_Group]
	entityGroupsCache *concurrent.Cache[uid.Id, []string]
}

// Init 初始化插件
func (r *_Router) Init(svcCtx service.Context) {
	log.Infof(svcCtx, "init addin %q", self.Name)

	r.svcCtx = svcCtx
	r.gate = gate.Using(r.svcCtx)

	if r.options.EtcdClient == nil {
		cli, err := etcdv3.New(r.configure())
		if err != nil {
			log.Panicf(svcCtx, "new etcd client failed, %s", err)
		}
		r.client = cli
	} else {
		r.client = r.options.EtcdClient
	}

	for _, ep := range r.client.Endpoints() {
		func() {
			ctx, cancel := context.WithTimeout(r.svcCtx, 3*time.Second)
			defer cancel()

			if _, err := r.client.Status(ctx, ep); err != nil {
				log.Panicf(r.svcCtx, "status etcd %q failed, %s", ep, err)
			}
		}()
	}

	r.groupCache = concurrent.NewCache[string, *_Group]()
	r.groupCache.OnDel(func(name string, group *_Group) { group.terminate() })

	r.entityGroupsCache = concurrent.NewCache[uid.Id, []string]()
	r.entityGroupsCache.AutoClean(r.svcCtx, 30*time.Second, 256)
}

// Shut 关闭插件
func (r *_Router) Shut(svcCtx service.Context) {
	log.Infof(svcCtx, "shut addin %q", self.Name)

	if r.options.EtcdClient == nil {
		if r.client != nil {
			r.client.Close()
		}
	}
}

// Mapping 路由映射
func (r *_Router) Mapping(entityId, sessionId uid.Id) (IMapping, error) {
	entity, ok := r.svcCtx.GetEntityManager().GetEntity(entityId)
	if !ok {
		return nil, ErrEntityNotFound
	}

	session, ok := r.gate.GetSession(sessionId)
	if !ok {
		return nil, ErrSessionNotFound
	}

	var ret IMapping
	var err error

	r.planning.AutoLock(func(planning *map[uid.Id]*_Mapping) {
		if _, ok := (*planning)[entityId]; ok {
			err = ErrEntityMapped
			return
		}

		if _, ok := (*planning)[sessionId]; ok {
			err = ErrSessionMapped
			return
		}

		ctx, cancel := context.WithCancel(r.svcCtx)

		mapping := &_Mapping{
			Context:   ctx,
			router:    r,
			terminate: cancel,
			entity:    entity,
			session:   session,
			cliAddr:   gate.CliDetails.DomainUnicast.Join(entity.GetId().String()),
		}
		ret = mapping

		(*planning)[entityId] = mapping
		(*planning)[sessionId] = mapping

		go mapping.mainLoop()
	})

	return ret, err
}

// CleanEntity 清理实体路由信息
func (r *_Router) CleanEntity(entityId uid.Id) {
	r.planning.AutoLock(func(planning *map[uid.Id]*_Mapping) {
		mapping, ok := (*planning)[entityId]
		if !ok {
			return
		}
		delete(*planning, entityId)

		if (*planning)[mapping.session.GetId()] == mapping {
			delete(*planning, mapping.session.GetId())
		}

		mapping.terminate()
	})
}

// CleanSession 清理会话路由信息
func (r *_Router) CleanSession(sessionId uid.Id) {
	r.planning.AutoLock(func(planning *map[uid.Id]*_Mapping) {
		mapping, ok := (*planning)[sessionId]
		if !ok {
			return
		}
		delete(*planning, sessionId)

		if (*planning)[mapping.entity.GetId()] == mapping {
			delete(*planning, mapping.entity.GetId())
		}

		mapping.terminate()
	})
}

// LookupEntity 查找实体
func (r *_Router) LookupEntity(sessionId uid.Id) (ec.ConcurrentEntity, string, bool) {
	mapping, ok := r.planning.Get(sessionId)
	if !ok {
		return nil, "", false
	}
	return mapping.entity, mapping.cliAddr, true
}

// LookupSession 查找会话
func (r *_Router) LookupSession(entityId uid.Id) (gate.ISession, bool) {
	mapping, ok := r.planning.Get(entityId)
	if !ok {
		return nil, false
	}
	return mapping.session, true
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
