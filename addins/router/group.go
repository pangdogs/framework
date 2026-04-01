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
	"errors"
	"fmt"
	"path"
	"slices"
	"sync"
	"sync/atomic"

	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/addins/gate"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/net/gtp/transport"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

// IGroup 路由组接口
type IGroup interface {
	// Name 获取名称
	Name() string
	// ClientAddr 获取客户端地址
	ClientAddr() string
	// KeepAliveContinuous 路由组持续保活
	KeepAliveContinuous(ctx context.Context) (async.Future, error)
	// KeepAliveOnce 路由组保活一次
	KeepAliveOnce(ctx context.Context) error
	// Deleted 等待路由组被删除
	Deleted() async.Future
	// Add 添加成员实体id
	Add(ctx context.Context, ids []uid.Id) error
	// Remove 移除成员实体id
	Remove(ctx context.Context, ids []uid.Id) error
	// List 列出成员实体id
	List() []uid.Id
	// DataIO 获取数据IO
	DataIO() IDataIO
	// EventIO 获取事件IO
	EventIO() IEventIO
}

type _Group struct {
	router          *_Router
	clientAddr      string
	leaseId         etcdv3.LeaseID
	createdRevision int64
	latestRevision  int64
	entities        atomic.Pointer[generic.SliceMap[uid.Id, int64]]
	io              _GroupIO
	expired         async.FutureVoid
	deleted         async.FutureVoid
	expireOnce      sync.Once
	deleteOnce      sync.Once
}

// Name 获取名称
func (g *_Group) Name() string {
	name, _ := gate.ClientDetails.DomainMulticast.Relative(g.clientAddr)
	return name
}

// ClientAddr 获取客户端地址
func (g *_Group) ClientAddr() string {
	return g.clientAddr
}

// KeepAliveContinuous 路由组持续保活
func (g *_Group) KeepAliveContinuous(ctx context.Context) (async.Future, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case <-g.router.ctx.Done():
		return async.Future{}, errors.New("router: router is terminating")
	default:
	}

	if !g.router.barrier.Join(1) {
		return async.Future{}, errors.New("router: router is terminating")
	}

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-ctx.Done():
		case <-g.expired:
		}
		cancel()
	}()

	keepAliveChan, err := g.router.client.KeepAlive(ctx, g.leaseId)
	if err != nil {
		cancel()
		g.router.barrier.Done()

		log.L(g.router.svcCtx).Error("keep alive group lease failed",
			zap.String("group_name", g.Name()),
			zap.String("group_addr", g.ClientAddr()),
			zap.Int64("lease_id", int64(g.leaseId)),
			zap.Error(err))
		return async.Future{}, fmt.Errorf("router: %w", err)
	}

	stopped := async.NewFutureVoid()

	go func() {
		defer func() {
			cancel()
			g.router.barrier.Done()
		}()

		for range keepAliveChan {
			log.L(g.router.svcCtx).Debug("keep alive group lease heartbeat ok",
				zap.String("group_name", g.Name()),
				zap.String("group_addr", g.ClientAddr()),
				zap.Int64("lease_id", int64(g.leaseId)))
		}

		log.L(g.router.svcCtx).Debug("keep alive group lease heartbeat closed",
			zap.String("group_name", g.Name()),
			zap.String("group_addr", g.ClientAddr()),
			zap.Int64("lease_id", int64(g.leaseId)))

		async.ReturnVoid(stopped)
	}()

	log.L(g.router.svcCtx).Debug("keep alive group lease ok",
		zap.String("group_name", g.Name()),
		zap.String("group_addr", g.ClientAddr()),
		zap.Int64("lease_id", int64(g.leaseId)))
	return stopped.Out(), nil
}

// KeepAliveOnce 路由组保活一次
func (g *_Group) KeepAliveOnce(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	_, err := g.router.client.KeepAliveOnce(ctx, g.leaseId)
	if err != nil {
		log.L(g.router.svcCtx).Error("keep alive group lease once failed",
			zap.String("group_name", g.Name()),
			zap.String("group_addr", g.ClientAddr()),
			zap.Int64("lease_id", int64(g.leaseId)),
			zap.Error(err))
		return fmt.Errorf("router: %w", err)
	}

	log.L(g.router.svcCtx).Debug("keep alive group lease once ok",
		zap.String("group_name", g.Name()),
		zap.String("group_addr", g.ClientAddr()),
		zap.Int64("lease_id", int64(g.leaseId)))
	return nil
}

// Deleted 等待路由组被删除
func (g *_Group) Deleted() async.Future {
	return g.deleted.Out()
}

// Add 添加成员实体id
func (g *_Group) Add(ctx context.Context, ids []uid.Id) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if len(ids) <= 0 {
		return nil
	}

	ops := make([]etcdv3.Op, 0, len(ids)*2)

	for _, id := range ids {
		ops = append(ops,
			etcdv3.OpPut(path.Join(g.router.groupEntitiesKeyPrefix, g.clientAddr, id.String()), "", etcdv3.WithLease(g.leaseId)),
			etcdv3.OpPut(path.Join(g.router.entityGroupsKeyPrefix, id.String(), g.clientAddr), "", etcdv3.WithLease(g.leaseId)),
		)
	}

	_, err := g.router.client.Txn(ctx).Then(ops...).Commit()
	if err != nil {
		log.L(g.router.svcCtx).Error("add group members failed",
			zap.String("group_name", g.Name()),
			zap.String("group_addr", g.ClientAddr()),
			zap.Any("entity_ids", ids),
			zap.Error(err))
		return fmt.Errorf("router: %w", err)
	}

	log.L(g.router.svcCtx).Info("group members added",
		zap.String("group_name", g.Name()),
		zap.String("group_addr", g.ClientAddr()),
		zap.Any("entity_ids", ids))
	return nil
}

// Remove 移除成员实体id
func (g *_Group) Remove(ctx context.Context, ids []uid.Id) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if len(ids) <= 0 {
		return nil
	}

	ops := make([]etcdv3.Op, 0, len(ids)*2)
	for _, id := range ids {
		ops = append(ops,
			etcdv3.OpDelete(path.Join(g.router.groupEntitiesKeyPrefix, g.clientAddr, id.String())),
			etcdv3.OpDelete(path.Join(g.router.entityGroupsKeyPrefix, id.String(), g.clientAddr)),
		)
	}

	_, err := g.router.client.Txn(ctx).Then(ops...).Commit()
	if err != nil {
		log.L(g.router.svcCtx).Error("remove group members failed",
			zap.String("group_name", g.Name()),
			zap.String("group_addr", g.ClientAddr()),
			zap.Any("entity_ids", ids),
			zap.Error(err))
		return fmt.Errorf("router: %w", err)
	}

	log.L(g.router.svcCtx).Info("group members removed",
		zap.String("group_name", g.Name()),
		zap.String("group_addr", g.ClientAddr()),
		zap.Any("entity_ids", ids))
	return nil
}

// List 列出成员实体id
func (g *_Group) List() []uid.Id {
	entities := g.getEntities()
	if len(entities) == 0 {
		return nil
	}
	return entities.Keys()
}

// DataIO 获取数据IO
func (g *_Group) DataIO() IDataIO {
	return (*_GroupDataIO)(&g.io)
}

// EventIO 获取事件IO
func (g *_Group) EventIO() IEventIO {
	return (*_GroupEventIO)(&g.io)
}

func (g *_Group) init(r *_Router, addr string, leaseId etcdv3.LeaseID, revision int64, ids []uid.Id) {
	g.router = r
	g.clientAddr = addr
	g.leaseId = leaseId
	g.createdRevision = revision
	g.latestRevision = revision
	g.expired = async.NewFutureVoid()
	g.deleted = async.NewFutureVoid()

	if len(ids) > 0 {
		entities := generic.NewSliceMap[uid.Id, int64]()
		for _, id := range ids {
			entities.Add(id, revision)
		}
		g.entities.Store(&entities)
	} else {
		g.entities.Store(nil)
	}

	g.io.init(g)
}

func (g *_Group) sendData(data []byte) error {
	var retErr []error

	g.getEntities().Each(func(id uid.Id, _ int64) {
		mapping, ok := g.router.Lookup(id)
		if !ok {
			return
		}
		if err := mapping.Session().DataIO().Send(data); err != nil {
			retErr = append(retErr, err)
		}
	})

	if len(retErr) > 0 {
		return errors.Join(retErr...)
	}

	return nil
}

func (g *_Group) sendEvent(event transport.IEvent) error {
	var retErr []error

	g.getEntities().Each(func(id uid.Id, _ int64) {
		mapping, ok := g.router.Lookup(id)
		if !ok {
			return
		}
		if err := mapping.Session().EventIO().Send(event); err != nil {
			retErr = append(retErr, err)
		}
	})

	if len(retErr) > 0 {
		return errors.Join(retErr...)
	}

	return nil
}

func (g *_Group) markExpired() {
	g.expireOnce.Do(func() {
		async.ReturnVoid(g.expired)
	})
}

func (g *_Group) markDeleted() {
	g.deleteOnce.Do(func() {
		async.ReturnVoid(g.deleted)
	})
}

func (g *_Group) watchingForChanges() {
	defer g.router.barrier.Done()

	go g.io.sendLoop()

	ctx, cancel := context.WithCancel(g.router.ctx)
	defer cancel()

	var deleted bool
	revision := g.createdRevision + 1
	groupIdKey := g.router.groupIdKey(g.clientAddr)
	groupEntitiesPrefix := g.router.groupEntitiesPrefix(g.clientAddr)
	groupIdWatchChan := g.router.client.Watch(ctx, groupIdKey, etcdv3.WithRev(revision))
	groupEntitiesWatchChan := g.router.client.Watch(ctx, groupEntitiesPrefix, etcdv3.WithPrefix(), etcdv3.WithRev(revision))

	log.L(g.router.svcCtx).Debug("watching for group changes started",
		zap.String("group_name", g.Name()),
		zap.String("group_addr", g.ClientAddr()),
		zap.Int64("revision", revision))

	for groupIdWatchChan != nil || groupEntitiesWatchChan != nil {
		select {
		case watchRsp, ok := <-groupIdWatchChan:
			if !ok {
				groupIdWatchChan = nil
				continue
			}
			if watchRsp.Canceled {
				log.L(g.router.svcCtx).Debug("watching for group changes canceled",
					zap.String("group_name", g.Name()),
					zap.String("group_addr", g.ClientAddr()),
					zap.Error(watchRsp.Err()))
				groupIdWatchChan = nil
				continue
			}
			if watchRsp.Err() != nil {
				log.L(g.router.svcCtx).Panic("watching for group changes unexpectedly interrupted",
					zap.String("group_name", g.Name()),
					zap.String("group_addr", g.ClientAddr()),
					zap.Error(watchRsp.Err()))
				cancel()
				groupIdWatchChan = nil
				groupEntitiesWatchChan = nil
				continue
			}

			for _, event := range watchRsp.Events {
				if event.Type != etcdv3.EventTypeDelete {
					continue
				}

				deleted = true
				g.latestRevision = max(g.latestRevision, watchRsp.Header.Revision)

				cancel()
				groupIdWatchChan = nil
				groupEntitiesWatchChan = nil
				break
			}

		case watchRsp, ok := <-groupEntitiesWatchChan:
			if !ok {
				groupEntitiesWatchChan = nil
				continue
			}
			if watchRsp.Canceled {
				log.L(g.router.svcCtx).Debug("watching for group changes canceled",
					zap.String("group_name", g.Name()),
					zap.String("group_addr", g.ClientAddr()),
					zap.Error(watchRsp.Err()))
				groupEntitiesWatchChan = nil
				continue
			}
			if watchRsp.Err() != nil {
				log.L(g.router.svcCtx).Panic("watching for group changes unexpectedly interrupted",
					zap.String("group_name", g.Name()),
					zap.String("group_addr", g.ClientAddr()),
					zap.Error(watchRsp.Err()))
				cancel()
				groupIdWatchChan = nil
				groupEntitiesWatchChan = nil
				continue
			}

			entities := g.getEntities()
			if len(entities) > 0 {
				entities = slices.Clone(entities)
			}

			for _, event := range watchRsp.Events {
				groupAddr, entityId, ok := g.router.parseGroupEntitiesKey(string(event.Kv.Key))
				if !ok || groupAddr != g.clientAddr {
					continue
				}

				switch event.Type {
				case etcdv3.EventTypePut:
					entities.Add(entityId, watchRsp.Header.Revision)

				case etcdv3.EventTypeDelete:
					entities.Delete(entityId)

				default:
					log.L(g.router.svcCtx).Warn("unknown group changes event type",
						zap.String("group_name", g.Name()),
						zap.String("group_addr", g.ClientAddr()),
						zap.String("type", event.Type.String()))
				}
			}

			g.storeEntities(entities)
			g.latestRevision = max(g.latestRevision, watchRsp.Header.Revision)
		}
	}

	g.router.uncacheGroup(g)
	g.markExpired()
	if deleted {
		g.markDeleted()
	}
	<-g.io.terminated

	log.L(g.router.svcCtx).Debug("watching for group changes stopped",
		zap.String("group_name", g.Name()),
		zap.String("group_addr", g.ClientAddr()),
		zap.Bool("deleted", deleted),
		zap.Int64("revision", g.latestRevision))
}

func (g *_Group) getEntities() generic.SliceMap[uid.Id, int64] {
	entities := g.entities.Load()
	if entities == nil {
		return nil
	}
	return *entities
}

func (g *_Group) storeEntities(entities generic.SliceMap[uid.Id, int64]) {
	if len(entities) == 0 {
		g.entities.Store(nil)
		return
	}

	cloned := slices.Clone(entities)
	g.entities.Store(&cloned)
}
