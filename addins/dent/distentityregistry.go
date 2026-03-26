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

package dent

import (
	"context"
	"crypto/tls"
	"math"
	"path"
	"time"

	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/event"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/addins/log"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

// IDistEntityRegistry 分布式实体注册支持接口（自动注册全局实体为分布式实体）
type IDistEntityRegistry interface {
	IDistEntityRegistryEventTab
}

func newDistEntityRegistry(settings ...option.Setting[DistEntityRegistryOptions]) IDistEntityRegistry {
	return &_DistEntityRegistry{
		options: option.New(With.Registry.Default(), settings...),
	}
}

type _DistEntityRegistry struct {
	distEntityRegistryEventTab
	rtCtx          runtime.Context
	options        DistEntityRegistryOptions
	client         *etcdv3.Client
	leaseId        etcdv3.LeaseID
	managedHandles [2]event.Handle
}

// Init 初始化插件
func (d *_DistEntityRegistry) Init(rtCtx runtime.Context) {
	log.L(rtCtx).Info("initializing add-in", zap.String("name", RegistryAddIn.Name))

	d.rtCtx = rtCtx
	d.distEntityRegistryEventTab.SetPanicHandling(rtCtx.AutoRecover(), rtCtx.ReportError())

	if d.options.EtcdClient == nil {
		cli, err := etcdv3.New(d.configure())
		if err != nil {
			log.L(rtCtx).Panic("new etcd client failed", log.JSON("config", d.configure()), zap.Error(err))
		}
		d.client = cli
	} else {
		d.client = d.options.EtcdClient
	}

	for _, ep := range d.client.Endpoints() {
		func() {
			ctx, cancel := context.WithTimeout(rtCtx, 3*time.Second)
			defer cancel()

			if _, err := d.client.Status(ctx, ep); err != nil {
				log.L(rtCtx).Panic("status etcd failed", zap.Any("endpoint", ep), zap.Error(err))
			}
		}()
	}

	// 申请租约
	grantRsp, err := d.client.Grant(rtCtx, int64(math.Ceil(d.options.RegistrationTTL.Seconds())))
	if err != nil {
		log.L(rtCtx).Panic("grant etcd lease failed", zap.Error(err))
	}
	keepAliveChan, err := d.client.KeepAlive(rtCtx, grantRsp.ID)
	if err != nil {
		log.L(rtCtx).Panic("keep alive etcd lease failed", zap.Error(err))
	}
	d.leaseId = grantRsp.ID

	log.L(rtCtx).Debug("grant etcd lease ok", zap.Int64("lease_id", int64(grantRsp.ID)))

	go func() {
		for range keepAliveChan {
			log.L(service.Current(rtCtx)).Debug("keep alive etcd lease heartbeat ok",
				zap.String("runtime_id", rtCtx.Id().String()),
				zap.Int64("lease_id", int64(d.leaseId)))
		}
		log.L(service.Current(rtCtx)).Debug("keep alive etcd lease heartbeat closed",
			zap.String("runtime_id", rtCtx.Id().String()),
			zap.Int64("lease_id", int64(d.leaseId)))
	}()

	// 刷新实体信息
	rtCtx.EntityManager().EachEntities(d.register)

	// 绑定事件
	d.managedHandles = [2]event.Handle{
		runtime.BindEventEntityManagerAddEntity(rtCtx.EntityManager(), d, 1000),
		runtime.BindEventEntityManagerRemoveEntity(rtCtx.EntityManager(), d, -1000),
	}
}

// Shut 关闭插件
func (d *_DistEntityRegistry) Shut(rtCtx runtime.Context) {
	log.L(rtCtx).Info("shutting down add-in", zap.String("name", RegistryAddIn.Name))

	// 解绑定事件钩子
	event.UnbindHandles(d.managedHandles[:])

	// 废除租约
	_, err := d.client.Revoke(context.Background(), d.leaseId)
	if err != nil {
		log.L(rtCtx).Error("revoke etcd lease failed", zap.Int64("lease_id", int64(d.leaseId)), zap.Error(err))
	}

	if d.options.EtcdClient == nil {
		if d.client != nil {
			d.client.Close()
		}
	}

	d.distEntityRegistryEventTab.SetEnabled(false)
}

// OnEntityManagerAddEntity 实体管理器添加实体
func (d *_DistEntityRegistry) OnEntityManagerAddEntity(entityMgr runtime.EntityManager, entity ec.Entity) {
	d.register(entity)
}

// OnEntityManagerRemoveEntity 实体管理器删除实体
func (d *_DistEntityRegistry) OnEntityManagerRemoveEntity(entityMgr runtime.EntityManager, entity ec.Entity) {
	d.deregister(entity)
}

func (d *_DistEntityRegistry) register(entity ec.Entity) {
	if entity.Scope() != ec.Scope_Global {
		return
	}

	key := d.newEntityKey(entity)

	_, err := d.client.Put(d.rtCtx, key, "", etcdv3.WithIgnoreValue(), etcdv3.WithLease(d.leaseId))
	if err != nil {
		log.L(d.rtCtx).Error("put etcd key with lease failed", zap.String("key", key), zap.Int64("lease_id", int64(d.leaseId)), zap.Error(err))
		return
	}
	log.L(d.rtCtx).Debug("put etcd key with lease ok", zap.String("key", key), zap.Int64("lease_id", int64(d.leaseId)))

	// 通知分布式实体上线
	_EmitEventDistEntityOnline(d, entity)
	return
}

func (d *_DistEntityRegistry) deregister(entity ec.Entity) {
	if entity.Scope() != ec.Scope_Global {
		return
	}

	select {
	case <-d.rtCtx.Done():
		break
	default:
		key := d.newEntityKey(entity)

		_, err := d.client.Delete(d.rtCtx, key)
		if err != nil {
			log.L(d.rtCtx).Error("delete etcd key failed", zap.String("key", key), zap.Int64("lease_id", int64(d.leaseId)), zap.Error(err))
		} else {
			log.L(d.rtCtx).Debug("delete etcd key ok", zap.String("key", key), zap.Int64("lease_id", int64(d.leaseId)))
		}
	}

	// 通知分布式实体下线
	_EmitEventDistEntityOffline(d, entity)
}

func (d *_DistEntityRegistry) configure() etcdv3.Config {
	if d.options.EtcdConfig != nil {
		return *d.options.EtcdConfig
	}

	config := etcdv3.Config{
		Endpoints:   d.options.CustomAddresses,
		Username:    d.options.CustomUsername,
		Password:    d.options.CustomPassword,
		DialTimeout: 3 * time.Second,
	}

	if d.options.CustomTLSConfig != nil {
		tlsConfig := d.options.CustomTLSConfig
		if tlsConfig == nil {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		config.TLS = tlsConfig
	}

	return config
}

func (d *_DistEntityRegistry) newEntityKey(entity ec.Entity) string {
	svcCtx := service.Current(d.rtCtx)
	return path.Join(d.options.KeyPrefix, entity.Id().String(), svcCtx.Name(), svcCtx.Id().String())
}
