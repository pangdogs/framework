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
	"bytes"
	"context"
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/addins/gate"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/net/gtp/transport"
	"git.golaxy.org/framework/utils/binaryutil"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"path"
	"slices"
	"strings"
	"sync"
)

// IGroup 分组接口
type IGroup interface {
	context.Context
	// GetName 获取分组名称
	GetName() string
	// GetAddr 获取分组地址
	GetAddr() string
	// Add 添加实体
	Add(ctx context.Context, entIds ...uid.Id) error
	// Remove 删除实体
	Remove(ctx context.Context, entIds ...uid.Id) error
	// Range 遍历所有实体
	Range(fun generic.Func1[uid.Id, bool])
	// Each 遍历所有实体
	Each(fun generic.Action1[uid.Id])
	// Count 获取实体数量
	Count() int
	// RefreshTTL 刷新TTL
	RefreshTTL(ctx context.Context) error
	// SendData 发送数据
	SendData(data []byte)
	// SendEvent 发送自定义事件
	SendEvent(event transport.IEvent)
	// SendDataChan 发送数据的channel
	SendDataChan() chan<- binaryutil.RecycleBytes
	// SendEventChan 发送自定义事件的channel
	SendEventChan() chan<- transport.IEvent
}

type _Group struct {
	context.Context
	sync.RWMutex
	terminate     context.CancelFunc
	router        *_Router
	groupKey      string
	leaseId       etcdv3.LeaseID
	revision      int64
	entities      []uid.Id
	sendDataChan  chan binaryutil.RecycleBytes
	sendEventChan chan transport.IEvent
}

// GetName 获取分组名称
func (g *_Group) GetName() string {
	name, _ := gate.CliDetails.DomainMulticast.Relative(g.GetAddr())
	return name
}

// GetAddr 获取分组地址
func (g *_Group) GetAddr() string {
	return strings.TrimPrefix(g.groupKey, g.router.options.GroupKeyPrefix)
}

// Add 添加实体
func (g *_Group) Add(ctx context.Context, entIds ...uid.Id) error {
	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case <-g.Done():
		return context.Canceled
	default:
		break
	}

	if len(entIds) <= 0 {
		return nil
	}

	opsPut := make([]etcdv3.Op, 0, len(entIds)*2)
	for _, entId := range entIds {
		opsPut = append(opsPut,
			etcdv3.OpPut(path.Join(g.groupKey, entId.String()), "", etcdv3.WithLease(g.leaseId)),
			etcdv3.OpPut(path.Join(g.router.options.EntityGroupsKeyPrefix, entId.String(), g.GetAddr()), "", etcdv3.WithLease(g.leaseId)),
		)
	}

	_, err := g.router.client.Txn(ctx).
		Then(opsPut...).
		Commit()

	return err
}

// Remove 删除实体
func (g *_Group) Remove(ctx context.Context, entIds ...uid.Id) error {
	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case <-g.Done():
		return context.Canceled
	default:
		break
	}

	if len(entIds) <= 0 {
		return nil
	}

	opsDel := make([]etcdv3.Op, 0, len(entIds)*2)
	for _, entId := range entIds {
		opsDel = append(opsDel,
			etcdv3.OpDelete(path.Join(g.groupKey, entId.String())),
			etcdv3.OpDelete(path.Join(g.router.options.EntityGroupsKeyPrefix, entId.String(), g.GetAddr())),
		)
	}

	_, err := g.router.client.Txn(ctx).
		Then(opsDel...).
		Commit()

	return err
}

// Range 遍历所有实体
func (g *_Group) Range(fun generic.Func1[uid.Id, bool]) {
	g.RLock()
	copied := slices.Clone(g.entities)
	g.RUnlock()

	for i := range copied {
		if !fun.UnsafeCall(copied[i]) {
			return
		}
	}
}

// Each 遍历所有实体
func (g *_Group) Each(fun generic.Action1[uid.Id]) {
	g.RLock()
	copied := slices.Clone(g.entities)
	g.RUnlock()

	for i := range copied {
		fun.UnsafeCall(copied[i])
	}
}

// Count 获取实体数量
func (g *_Group) Count() int {
	g.RLock()
	defer g.Unlock()
	return len(g.entities)
}

// RefreshTTL 刷新TTL
func (g *_Group) RefreshTTL(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case <-g.Done():
		return context.Canceled
	default:
		break
	}

	_, err := g.router.client.KeepAliveOnce(ctx, g.leaseId)
	return err
}

// SendData 发送数据
func (g *_Group) SendData(data []byte) {
	g.RLock()
	defer g.RUnlock()

	select {
	case <-g.Done():
		return
	default:
		break
	}

	for i := range g.entities {
		entId := g.entities[i]

		g.router.svcCtx.CallVoidAsync(entId, func(entity ec.Entity, _ ...any) {
			session, ok := g.router.LookupSession(entity.GetId())
			if !ok {
				return
			}

			err := session.SendData(data)
			if err != nil {
				log.Errorf(g.router.svcCtx, "group %q send data to session %q remote %q failed, %s", g.GetName(), session.GetId(), session.GetRemoteAddr(), err)
			}
		})
	}
}

// SendEvent 发送自定义事件
func (g *_Group) SendEvent(event transport.IEvent) {
	g.RLock()
	defer g.RUnlock()

	select {
	case <-g.Done():
		return
	default:
		break
	}

	for i := range g.entities {
		entId := g.entities[i]

		g.router.svcCtx.CallVoidAsync(entId, func(entity ec.Entity, _ ...any) {
			session, ok := g.router.LookupSession(entity.GetId())
			if !ok {
				return
			}

			err := session.SendEvent(event)
			if err != nil {
				log.Errorf(g.router.svcCtx, "group %q send event(%d) to session %q remote %q failed, %s", g.GetName(), event.Msg.MsgId(), session.GetId(), session.GetRemoteAddr(), err)
			}
		})
	}
}

// SendDataChan 发送数据的channel
func (g *_Group) SendDataChan() chan<- binaryutil.RecycleBytes {
	if g.sendDataChan == nil {
		log.Panicf(g.router.svcCtx, "group %q send data channel size less equal 0, can't be used", g.GetName())
	}
	return g.sendDataChan
}

// SendEventChan 发送自定义事件的channel
func (g *_Group) SendEventChan() chan<- transport.IEvent {
	if g.sendEventChan == nil {
		log.Panicf(g.router.svcCtx, "group %q send event channel size less equal 0, can't be used", g.GetName())
	}
	return g.sendEventChan
}

func (g *_Group) mainLoop() {
	ctx, cancel := context.WithCancel(g)

	if g.router.options.GroupAutoRefreshTTL {
		rspChan, err := g.router.client.KeepAlive(ctx, g.leaseId)
		if err != nil {
			log.Errorf(g.router.svcCtx, "keep alive groupKey %q lease %q failed, %s", g.groupKey, g.leaseId, err)
			goto watch
		}

		go func() {
			for range rspChan {
				log.Debugf(g.router.svcCtx, "refresh groupKey %q ttl success", g.groupKey)
			}
		}()
	}

	if g.sendDataChan != nil {
		go func() {
			defer func() {
				for bs := range g.sendDataChan {
					bs.Release()
				}
			}()
			for {
				select {
				case bs := <-g.sendDataChan:
					g.SendData(bytes.Clone(bs.Data()))
					bs.Release()
				case <-g.Done():
					return
				}
			}
		}()
	}

	if g.sendEventChan != nil {
		go func() {
			for {
				select {
				case event := <-g.sendEventChan:
					g.SendEvent(event)
				case <-g.Done():
					return
				}
			}
		}()
	}

watch:
	watchChan := g.router.client.Watch(ctx, g.groupKey, etcdv3.WithRev(g.revision), etcdv3.WithPrefix(), etcdv3.WithIgnoreValue())

	log.Debugf(g.router.svcCtx, "start watch groupKey %q", g.groupKey)

	for watchRsp := range watchChan {
		if watchRsp.Canceled {
			log.Debugf(g.router.svcCtx, "stop watch groupKey %q", g.groupKey)
			goto end
		}
		if watchRsp.Err() != nil {
			log.Errorf(g.router.svcCtx, "interrupt watch groupKey %q, %s", g.groupKey, watchRsp.Err())
			goto end
		}

		g.Lock()
		for _, event := range watchRsp.Events {
			switch event.Type {
			case etcdv3.EventTypePut:
				key := string(event.Kv.Key)
				if key == g.groupKey {
					continue
				}

				entId := uid.From(path.Base(key))

				if !slices.Contains(g.entities, entId) {
					g.entities = append(g.entities, entId)
				}

				g.router.entityGroupsCache.Del(entId, watchRsp.Header.Revision)

			case etcdv3.EventTypeDelete:
				key := string(event.Kv.Key)
				if key == g.groupKey {
					cancel()
					continue
				}

				entId := uid.From(path.Base(key))

				idx := slices.Index(g.entities, entId)
				if idx >= 0 {
					g.entities = slices.Delete(g.entities, idx, idx+1)
				}

				g.router.entityGroupsCache.Del(entId, watchRsp.Header.Revision)

			default:
				log.Errorf(g.router.svcCtx, "unknown event type %q", event.Type)
				continue
			}
		}

		if g.revision < watchRsp.Header.Revision {
			g.revision = watchRsp.Header.Revision
		}
		g.Unlock()
	}

end:
	g.RLock()
	for _, entId := range g.entities {
		g.router.entityGroupsCache.Del(entId, g.revision)
	}
	g.router.groupCache.Del(g.GetName(), g.revision)
	g.RUnlock()
}
