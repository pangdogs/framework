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
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/net/gtp/transport"
	"git.golaxy.org/framework/plugins/gate"
	"git.golaxy.org/framework/utils/binaryutil"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"math"
	"path"
	"strconv"
	"strings"
)

// AddGroup 添加分组
func (r *_Router) AddGroup(ctx context.Context, name string) (IGroup, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	lgr, err := r.client.Grant(ctx, int64(math.Ceil(r.options.GroupTTL.Seconds())))
	if err != nil {
		return nil, err
	}
	leaseId := lgr.ID

	groupAddr := gate.CliDetails.DomainMulticast.Join(name)
	groupKey := path.Join(r.options.GroupKeyPrefix, groupAddr)

	tr, err := r.client.Txn(ctx).
		If(etcdv3.Compare(etcdv3.Version(groupKey), "=", 0)).
		Then(
			etcdv3.OpPut(groupKey, strconv.Itoa(int(leaseId)), etcdv3.WithLease(leaseId)),
		).
		Else(
			etcdv3.OpGet(groupKey),
			etcdv3.OpGet(groupKey+"/",
				etcdv3.WithPrefix(),
				etcdv3.WithSort(etcdv3.SortByModRevision, etcdv3.SortDescend),
				etcdv3.WithIgnoreValue(),
			),
		).
		Commit()
	if err != nil {
		return nil, err
	}

	var entIds []uid.Id

	if !tr.Succeeded {
		if len(tr.Responses[0].GetResponseRange().Kvs) <= 0 {
			return nil, errors.New("missing groupKey")
		}

		l, err := strconv.Atoi(string(tr.Responses[0].GetResponseRange().Kvs[0].Value))
		if err != nil {
			return nil, errors.New("missing groupKey leaseId")
		}
		leaseId = etcdv3.LeaseID(l)

		entIds = make([]uid.Id, 0, len(tr.Responses[1].GetResponseRange().Kvs))
		for _, kv := range tr.Responses[1].GetResponseRange().Kvs {
			entIds = append(entIds, uid.From(path.Base(string(kv.Key))))
		}
	}

	group := r.newGroup(groupKey, leaseId, tr.Header.Revision, entIds)

	cached := r.groupCache.Set(group.GetName(), group, tr.Header.Revision, 0)
	if cached == group {
		go group.mainLoop()
	}

	return cached, nil
}

// DeleteGroup 删除分组
func (r *_Router) DeleteGroup(ctx context.Context, name string) {
	if ctx == nil {
		ctx = context.Background()
	}

	groupAddr := gate.CliDetails.DomainMulticast.Join(name)
	groupKey := path.Join(r.options.GroupKeyPrefix, groupAddr)

	gr, err := r.client.Get(ctx, groupKey)
	if err != nil {
		return
	}

	if len(gr.Kvs) <= 0 {
		return
	}

	l, err := strconv.Atoi(string(gr.Kvs[0].Value))
	if err != nil {
		return
	}
	leaseId := etcdv3.LeaseID(l)

	_, err = r.client.Revoke(context.Background(), leaseId)
	if err != nil {
		return
	}
}

// GetGroup 查询分组
func (r *_Router) GetGroup(ctx context.Context, name string) (IGroup, bool) {
	if ctx == nil {
		ctx = context.Background()
	}

	group, ok := r.groupCache.Get(name)
	if ok {
		return group, true
	}

	groupAddr := gate.CliDetails.DomainMulticast.Join(name)
	groupKey := path.Join(r.options.GroupKeyPrefix, groupAddr)

	tr, err := r.client.Txn(ctx).
		If(etcdv3.Compare(etcdv3.Version(groupKey), "!=", 0)).
		Then(
			etcdv3.OpGet(groupKey),
			etcdv3.OpGet(groupKey+"/",
				etcdv3.WithPrefix(),
				etcdv3.WithSort(etcdv3.SortByModRevision, etcdv3.SortDescend),
				etcdv3.WithIgnoreValue(),
			),
		).
		Commit()
	if err != nil {
		return nil, false
	}

	if !tr.Succeeded || len(tr.Responses[0].GetResponseRange().Kvs) <= 0 {
		return nil, false
	}

	l, err := strconv.Atoi(string(tr.Responses[0].GetResponseRange().Kvs[0].Value))
	if err != nil {
		return nil, false
	}
	leaseId := etcdv3.LeaseID(l)

	entIds := make([]uid.Id, 0, len(tr.Responses[1].GetResponseRange().Kvs))
	for _, kv := range tr.Responses[1].GetResponseRange().Kvs {
		entIds = append(entIds, uid.From(path.Base(string(kv.Key))))
	}

	group = r.newGroup(groupKey, leaseId, tr.Header.Revision, entIds)

	cached := r.groupCache.Set(group.GetName(), group, tr.Header.Revision, 0)
	if cached == group {
		go group.mainLoop()
	}

	return cached, true
}

// GetGroupByAddr 使用分组地址查询分组
func (r *_Router) GetGroupByAddr(ctx context.Context, addr string) (IGroup, bool) {
	name, ok := gate.CliDetails.DomainMulticast.Relative(addr)
	if !ok {
		return nil, false
	}
	return r.GetGroup(ctx, name)
}

// RangeGroups 遍历包含实体的所有分组
func (r *_Router) RangeGroups(ctx context.Context, entityId uid.Id, fun generic.Func1[IGroup, bool]) {
	if ctx == nil {
		ctx = context.Background()
	}

	if _, ok := r.servCtx.GetEntityMgr().GetEntity(entityId); !ok {
		return
	}

	for _, groupAddr := range r.getEntityGroupAddrs(ctx, entityId) {
		if group, ok := r.GetGroupByAddr(ctx, groupAddr); ok {
			if !fun.Exec(group) {
				return
			}
		}
	}
}

// EachGroups 遍历包含实体的所有分组
func (r *_Router) EachGroups(ctx context.Context, entityId uid.Id, fun generic.Action1[IGroup]) {
	if ctx == nil {
		ctx = context.Background()
	}

	if _, ok := r.servCtx.GetEntityMgr().GetEntity(entityId); !ok {
		return
	}

	for _, groupAddr := range r.getEntityGroupAddrs(ctx, entityId) {
		if group, ok := r.GetGroupByAddr(ctx, groupAddr); ok {
			fun.Exec(group)
		}
	}
}

func (r *_Router) newGroup(groupKey string, leaseId etcdv3.LeaseID, revision int64, entIds []uid.Id) *_Group {
	group := &_Group{
		router:   r,
		groupKey: groupKey,
		leaseId:  leaseId,
		revision: revision,
		entities: entIds,
	}

	group.Context, group.terminate = context.WithCancel(r.servCtx)

	if r.options.GroupSendDataChanSize > 0 {
		group.sendDataChan = make(chan binaryutil.RecycleBytes, r.options.GroupSendDataChanSize)
	}

	if r.options.GroupSendEventChanSize > 0 {
		group.sendEventChan = make(chan transport.IEvent, r.options.GroupSendEventChanSize)
	}

	return group
}

func (r *_Router) getEntityGroupAddrs(ctx context.Context, entityId uid.Id) []string {
	groupAddrs, ok := r.entityGroupsCache.Get(entityId)
	if !ok {
		gr, err := r.client.Get(ctx, path.Join(r.options.EntityGroupsKeyPrefix, entityId.String())+"/",
			etcdv3.WithPrefix(),
			etcdv3.WithSort(etcdv3.SortByModRevision, etcdv3.SortDescend),
			etcdv3.WithIgnoreValue())
		if err != nil {
			return nil
		}

		if len(gr.Kvs) <= 0 {
			return nil
		}

		groupAddrs = make([]string, 0, len(gr.Kvs))

		for _, kv := range gr.Kvs {
			groupAddrs = append(groupAddrs, strings.TrimPrefix(string(kv.Key), r.options.GroupKeyPrefix))
		}

		groupAddrs = r.entityGroupsCache.Set(entityId, groupAddrs, gr.Header.Revision, r.options.EntityGroupsCacheTTL)
	}
	return groupAddrs
}
