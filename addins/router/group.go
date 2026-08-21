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

// IGroup 表示由 ETCD 租约维护、可向成员实体对应会话组播的路由组。
type IGroup interface {
	// Name 返回组播地址对应的逻辑组名。
	Name() string
	// ClientAddr 返回客户端使用的组播地址。
	ClientAddr() string
	// KeepAliveContinuous 持续刷新组租约，直到 ctx 取消或组过期。
	// 返回的 Signal 在保活停止后完成。
	KeepAliveContinuous(ctx context.Context) (async.Signal, error)
	// KeepAliveOnce 立即刷新一次组租约。
	KeepAliveOnce(ctx context.Context) error
	// Deleted 返回仅在组记录被显式删除时完成的 Signal。
	Deleted() async.Signal
	// Add 将实体 ID 加入组；重复成员不会产生额外条目。
	Add(ctx context.Context, ids []uid.ID) error
	// Remove 从组中移除实体 ID；不存在的成员会被忽略。
	Remove(ctx context.Context, ids []uid.ID) error
	// List 返回当前成员实体 ID 的快照。
	List() []uid.ID
	// DataIO 返回向当前可达成员会话发送原始数据的 I/O 门面。
	DataIO() IDataIO
	// EventIO 返回向当前可达成员会话发送传输事件的 I/O 门面。
	EventIO() IEventIO
}

type _Group struct {
	router          *_Router
	clientAddr      string
	leaseID         etcdv3.LeaseID
	createdRevision int64
	latestRevision  int64
	entities        atomic.Pointer[generic.SliceMap[uid.ID, int64]]
	io              _GroupIO
	expired         async.Completer
	deleted         async.Completer
	expireOnce      sync.Once
	deleteOnce      sync.Once
}

// Name 返回组播地址对应的逻辑组名。
func (g *_Group) Name() string {
	name, _ := gate.ClientDetails.DomainMulticast.Relative(g.clientAddr)
	return name
}

// ClientAddr 返回客户端使用的组播地址。
func (g *_Group) ClientAddr() string {
	return g.clientAddr
}

// KeepAliveContinuous 持续刷新组租约，直到 ctx 取消、路由器停止或组过期。
func (g *_Group) KeepAliveContinuous(ctx context.Context) (async.Signal, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case <-g.router.scope.Context().Done():
		return async.Signal{}, errors.New("router: router is terminating")
	default:
	}

	if !g.router.barrier.Join(1) {
		return async.Signal{}, errors.New("router: router is terminating")
	}

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-ctx.Done():
		case <-g.expired.Signal().Done():
		}
		cancel()
	}()

	keepAliveChan, err := g.router.client.KeepAlive(ctx, g.leaseID)
	if err != nil {
		cancel()
		g.router.barrier.Done()

		log.L(g.router.svcCtx).Error("keep alive group lease failed",
			zap.String("group_name", g.Name()),
			zap.String("group_addr", g.ClientAddr()),
			zap.Int64("lease_id", int64(g.leaseID)),
			zap.Error(err))
		return async.Signal{}, fmt.Errorf("router: %w", err)
	}

	stopped, stoppedSignal := async.NewSignal()

	go func() {
		defer g.router.barrier.Done()
		defer cancel()

		for range keepAliveChan {
			log.L(g.router.svcCtx).Debug("keep alive group lease heartbeat ok",
				zap.String("group_name", g.Name()),
				zap.String("group_addr", g.ClientAddr()),
				zap.Int64("lease_id", int64(g.leaseID)))
		}

		log.L(g.router.svcCtx).Debug("keep alive group lease heartbeat closed",
			zap.String("group_name", g.Name()),
			zap.String("group_addr", g.ClientAddr()),
			zap.Int64("lease_id", int64(g.leaseID)))

		stopped.Complete()
	}()

	log.L(g.router.svcCtx).Debug("keep alive group lease ok",
		zap.String("group_name", g.Name()),
		zap.String("group_addr", g.ClientAddr()),
		zap.Int64("lease_id", int64(g.leaseID)))
	return stoppedSignal, nil
}

// KeepAliveOnce 立即刷新一次组租约。
func (g *_Group) KeepAliveOnce(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	_, err := g.router.client.KeepAliveOnce(ctx, g.leaseID)
	if err != nil {
		log.L(g.router.svcCtx).Error("keep alive group lease once failed",
			zap.String("group_name", g.Name()),
			zap.String("group_addr", g.ClientAddr()),
			zap.Int64("lease_id", int64(g.leaseID)),
			zap.Error(err))
		return fmt.Errorf("router: %w", err)
	}

	log.L(g.router.svcCtx).Debug("keep alive group lease once ok",
		zap.String("group_name", g.Name()),
		zap.String("group_addr", g.ClientAddr()),
		zap.Int64("lease_id", int64(g.leaseID)))
	return nil
}

// Deleted 返回仅在组记录被显式删除时完成的 Signal。
func (g *_Group) Deleted() async.Signal {
	return g.deleted.Signal()
}

// Add 在同一 ETCD 事务中将实体 ID 加入组及其反向索引。
func (g *_Group) Add(ctx context.Context, ids []uid.ID) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if len(ids) <= 0 {
		return nil
	}

	ops := make([]etcdv3.Op, 0, len(ids)*2)

	for _, id := range ids {
		ops = append(ops,
			etcdv3.OpPut(path.Join(g.router.groupEntitiesKeyPrefix, g.clientAddr, id.String()), "", etcdv3.WithLease(g.leaseID)),
			etcdv3.OpPut(path.Join(g.router.entityGroupsKeyPrefix, id.String(), g.clientAddr), "", etcdv3.WithLease(g.leaseID)),
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

// Remove 在同一 ETCD 事务中移除实体 ID 及其反向索引。
func (g *_Group) Remove(ctx context.Context, ids []uid.ID) error {
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

// List 返回当前成员实体 ID 的快照。
func (g *_Group) List() []uid.ID {
	entities := g.getEntities()
	if len(entities) == 0 {
		return nil
	}
	return entities.Keys()
}

// DataIO 返回向当前可达成员会话发送原始数据的 I/O 门面。
func (g *_Group) DataIO() IDataIO {
	return (*_GroupDataIO)(&g.io)
}

// EventIO 返回向当前可达成员会话发送传输事件的 I/O 门面。
func (g *_Group) EventIO() IEventIO {
	return (*_GroupEventIO)(&g.io)
}

func (g *_Group) init(r *_Router, addr string, leaseID etcdv3.LeaseID, revision int64, ids []uid.ID) {
	g.router = r
	g.clientAddr = addr
	g.leaseID = leaseID
	g.createdRevision = revision
	g.latestRevision = revision
	g.expired, _ = async.NewSignal()
	g.deleted, _ = async.NewSignal()

	if len(ids) > 0 {
		entities := generic.NewSliceMap[uid.ID, int64]()
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

	g.getEntities().Each(func(id uid.ID, _ int64) {
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

	g.getEntities().Each(func(id uid.ID, _ int64) {
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
		g.expired.Complete()
	})
}

func (g *_Group) markDeleted() {
	g.deleteOnce.Do(func() {
		g.deleted.Complete()
	})
}

func (g *_Group) watchingForChanges() {
	defer g.router.barrier.Done()

	go g.io.sendLoop()

	ctx, cancel := context.WithCancel(g.router.scope.Context())
	defer cancel()

	var deleted bool
	revision := g.createdRevision + 1
	groupIDKey := g.router.groupIDKey(g.clientAddr)
	groupEntitiesPrefix := g.router.groupEntitiesPrefix(g.clientAddr)
	groupIDWatchChan := g.router.client.Watch(ctx, groupIDKey, etcdv3.WithRev(revision))
	groupEntitiesWatchChan := g.router.client.Watch(ctx, groupEntitiesPrefix, etcdv3.WithPrefix(), etcdv3.WithRev(revision))

	log.L(g.router.svcCtx).Debug("watching for group changes started",
		zap.String("group_name", g.Name()),
		zap.String("group_addr", g.ClientAddr()),
		zap.Int64("revision", revision))

	for groupIDWatchChan != nil || groupEntitiesWatchChan != nil {
		select {
		case watchRsp, ok := <-groupIDWatchChan:
			if !ok {
				groupIDWatchChan = nil
				continue
			}
			if watchRsp.Canceled {
				log.L(g.router.svcCtx).Debug("watching for group changes canceled",
					zap.String("group_name", g.Name()),
					zap.String("group_addr", g.ClientAddr()),
					zap.Error(watchRsp.Err()))
				groupIDWatchChan = nil
				continue
			}
			if watchRsp.Err() != nil {
				log.L(g.router.svcCtx).Panic("watching for group changes unexpectedly interrupted",
					zap.String("group_name", g.Name()),
					zap.String("group_addr", g.ClientAddr()),
					zap.Error(watchRsp.Err()))
				cancel()
				groupIDWatchChan = nil
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
				groupIDWatchChan = nil
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
				groupIDWatchChan = nil
				groupEntitiesWatchChan = nil
				continue
			}

			entities := g.getEntities()
			if len(entities) > 0 {
				entities = slices.Clone(entities)
			}

			for _, event := range watchRsp.Events {
				groupAddr, entityID, ok := g.router.parseGroupEntitiesKey(string(event.Kv.Key))
				if !ok || groupAddr != g.clientAddr {
					continue
				}

				switch event.Type {
				case etcdv3.EventTypePut:
					entities.Add(entityID, watchRsp.Header.Revision)

				case etcdv3.EventTypeDelete:
					entities.Delete(entityID)

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
	<-g.io.terminated.Signal().Done()

	log.L(g.router.svcCtx).Debug("watching for group changes stopped",
		zap.String("group_name", g.Name()),
		zap.String("group_addr", g.ClientAddr()),
		zap.Bool("deleted", deleted),
		zap.Int64("revision", g.latestRevision))
}

func (g *_Group) getEntities() generic.SliceMap[uid.ID, int64] {
	entities := g.entities.Load()
	if entities == nil {
		return nil
	}
	return *entities
}

func (g *_Group) storeEntities(entities generic.SliceMap[uid.ID, int64]) {
	if len(entities) == 0 {
		g.entities.Store(nil)
		return
	}

	cloned := slices.Clone(entities)
	g.entities.Store(&cloned)
}
