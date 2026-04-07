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
	"math"
	"path"
	"strings"
	"time"

	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/addins/gate"
	"git.golaxy.org/framework/addins/log"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

// AddGroup 创建路由组
func (r *_Router) AddGroup(ctx context.Context, name string, ids []uid.Id, ttl time.Duration) (IGroup, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case <-r.ctx.Done():
		return nil, errors.New("router: router is terminating")
	default:
	}

	if !r.barrier.Join(1) {
		return nil, errors.New("router: router is terminating")
	}

	groupAddr := gate.ClientDetails.DomainMulticast.Join(name)
	grantRsp, err := r.client.Grant(ctx, int64(math.Ceil(max(ttl.Seconds(), 3))))
	if err != nil {
		log.L(r.svcCtx).Error("grant group lease failed",
			zap.String("group_name", name),
			zap.String("group_addr", groupAddr),
			zap.Stringers("entity_ids", ids),
			zap.Duration("ttl", ttl),
			zap.Error(err))
		r.barrier.Done()
		return nil, fmt.Errorf("router: %w", err)
	}
	leaseId := grantRsp.ID

	groupIdKey := r.groupIdKey(groupAddr)
	ops := make([]etcdv3.Op, 0, 1+len(ids)*2)
	ops = append(ops, etcdv3.OpPut(groupIdKey, "", etcdv3.WithLease(leaseId)))

	for _, id := range ids {
		ops = append(ops,
			etcdv3.OpPut(path.Join(r.groupEntitiesKeyPrefix, groupAddr, id.String()), "", etcdv3.WithLease(leaseId)),
			etcdv3.OpPut(path.Join(r.entityGroupsKeyPrefix, id.String(), groupAddr), "", etcdv3.WithLease(leaseId)),
		)
	}

	tr, err := r.client.Txn(ctx).
		If(etcdv3.Compare(etcdv3.Version(groupIdKey), "=", 0)).
		Then(ops...).
		Else(etcdv3.OpGet(groupIdKey)).
		Commit()
	if err != nil {
		r.revokeGroupLease(context.Background(), leaseId)

		log.L(r.svcCtx).Error("put group keys failed",
			zap.String("group_name", name),
			zap.String("group_addr", groupAddr),
			zap.Stringers("entity_ids", ids),
			zap.Duration("ttl", ttl),
			zap.Int64("lease_id", int64(leaseId)),
			zap.Error(err))

		r.barrier.Done()
		return nil, fmt.Errorf("router: %w", err)
	}
	if !tr.Succeeded {
		r.revokeGroupLease(context.Background(), leaseId)

		log.L(r.svcCtx).Error("put group keys failed",
			zap.String("group_name", name),
			zap.String("group_addr", groupAddr),
			zap.Stringers("entity_ids", ids),
			zap.Duration("ttl", ttl),
			zap.Int64("lease_id", int64(leaseId)),
			zap.Error(ErrGroupExists))

		r.barrier.Done()
		return nil, ErrGroupExists
	}

	group, loaded := r.cacheGroup(name, groupAddr, leaseId, tr.Header.Revision, ids)
	if loaded {
		r.barrier.Done()
		return group, nil
	}

	return group, nil
}

// DeleteGroup 删除路由组
func (r *_Router) DeleteGroup(ctx context.Context, name string) {
	if ctx == nil {
		ctx = context.Background()
	}

	if cached, ok := r.groups.Load(name); ok {
		group := cached.(*_Group)

		if err := r.revokeGroupLease(ctx, group.leaseId); err != nil {
			log.L(r.svcCtx).Error("revoke group lease failed",
				zap.String("group_name", name),
				zap.String("group_addr", group.ClientAddr()),
				zap.Int64("lease_id", int64(group.leaseId)),
				zap.Error(err))
			return
		}

		r.uncacheGroup(group)
		group.markExpired()
		group.markDeleted()

		log.L(r.svcCtx).Info("delete group keys, local cache deleted",
			zap.String("group_name", name),
			zap.String("group_addr", group.ClientAddr()),
			zap.Int64("lease_id", int64(group.leaseId)))
		return
	}

	groupAddr := gate.ClientDetails.DomainMulticast.Join(name)
	rsp, err := r.client.Get(ctx, r.groupIdKey(groupAddr), etcdv3.WithKeysOnly())
	if err != nil {
		log.L(r.svcCtx).Error("get group keys lease failed",
			zap.String("group_name", name),
			zap.String("group_addr", groupAddr),
			zap.Error(err))
		return
	}
	if len(rsp.Kvs) <= 0 {
		log.L(r.svcCtx).Info("delete group keys, group not found",
			zap.String("group_name", name),
			zap.String("group_addr", groupAddr),
		)
		return
	}

	leaseId := etcdv3.LeaseID(rsp.Kvs[0].Lease)
	if err := r.revokeGroupLease(ctx, leaseId); err != nil {
		log.L(r.svcCtx).Error("revoke group lease failed",
			zap.String("group_name", name),
			zap.String("group_addr", groupAddr),
			zap.Int64("lease_id", int64(leaseId)),
			zap.Error(err))
		return
	}

	log.L(r.svcCtx).Info("delete group keys, not cached locally",
		zap.String("group_name", name),
		zap.String("group_addr", groupAddr),
		zap.Int64("lease_id", int64(leaseId)))
}

// GetGroupByName 使用名称查询路由组
func (r *_Router) GetGroupByName(ctx context.Context, name string) (IGroup, bool) {
	return r.getGroupByName(ctx, name)
}

// GetGroupByAddr 使用地址查询路由组
func (r *_Router) GetGroupByAddr(ctx context.Context, addr string) (IGroup, bool) {
	name, ok := gate.ClientDetails.DomainMulticast.Relative(addr)
	if !ok {
		return nil, false
	}
	return r.getGroupByName(ctx, name)
}

// GetGroupsByEntity 查询实体所属的所有路由组
func (r *_Router) GetGroupsByEntity(ctx context.Context, entityId uid.Id) []IGroup {
	if ctx == nil {
		ctx = context.Background()
	}

	rsp, err := r.client.Get(ctx, r.entityGroupsPrefix(entityId), etcdv3.WithPrefix(), etcdv3.WithKeysOnly())
	if err != nil {
		log.L(r.svcCtx).Error("get entity groups failed",
			zap.Stringer("entity_id", entityId),
			zap.Error(err))
		return nil
	}
	if len(rsp.Kvs) <= 0 {
		return nil
	}

	groups := make([]IGroup, 0, len(rsp.Kvs))
	seen := make(map[string]struct{}, len(rsp.Kvs))

	for _, kv := range rsp.Kvs {
		_, groupAddr, ok := r.parseEntityGroupsKey(string(kv.Key))
		if !ok {
			log.L(r.svcCtx).Warn("invalid entity groups key", zap.ByteString("key", kv.Key))
			continue
		}
		if _, ok := seen[groupAddr]; ok {
			log.L(r.svcCtx).Warn("duplicate entity groups key", zap.ByteString("key", kv.Key))
			continue
		}
		seen[groupAddr] = struct{}{}

		group, ok := r.GetGroupByAddr(ctx, groupAddr)
		if !ok {
			log.L(r.svcCtx).Warn("group not found for entity",
				zap.Stringer("entity_id", entityId),
				zap.String("group_addr", groupAddr))
			continue
		}

		groups = append(groups, group)
	}

	if len(groups) <= 0 {
		return nil
	}
	return groups
}

func (r *_Router) getGroupByName(ctx context.Context, groupName string) (IGroup, bool) {
	if ctx == nil {
		ctx = context.Background()
	}

	if cached, ok := r.groups.Load(groupName); ok {
		return cached.(*_Group), true
	}

	select {
	case <-r.ctx.Done():
		return nil, false
	default:
	}

	if !r.barrier.Join(1) {
		return nil, false
	}

	groupAddr := gate.ClientDetails.DomainMulticast.Join(groupName)
	groupIdKey := r.groupIdKey(groupAddr)
	groupEntitiesPrefix := r.groupEntitiesPrefix(groupAddr)

	tr, err := r.client.Txn(ctx).
		If(etcdv3.Compare(etcdv3.Version(groupIdKey), "!=", 0)).
		Then(
			etcdv3.OpGet(groupIdKey),
			etcdv3.OpGet(groupEntitiesPrefix, etcdv3.WithPrefix()),
		).
		Commit()
	if err != nil {
		log.L(r.svcCtx).Error("get group keys failed",
			zap.String("group_name", groupName),
			zap.String("group_addr", groupAddr),
			zap.Error(err))
		r.barrier.Done()
		return nil, false
	}
	if !tr.Succeeded || len(tr.Responses) < 2 {
		r.barrier.Done()
		return nil, false
	}

	groupIdRsp := tr.Responses[0].GetResponseRange()
	if groupIdRsp == nil || len(groupIdRsp.Kvs) <= 0 {
		r.barrier.Done()
		return nil, false
	}

	var ids []uid.Id
	entityRsp := tr.Responses[1].GetResponseRange()
	if entityRsp != nil {
		ids = make([]uid.Id, 0, len(entityRsp.Kvs))
		for _, kv := range entityRsp.Kvs {
			_, entityId, ok := r.parseGroupEntitiesKey(string(kv.Key))
			if ok {
				ids = append(ids, entityId)
			}
		}
	}

	group, loaded := r.cacheGroup(groupName, groupAddr, etcdv3.LeaseID(groupIdRsp.Kvs[0].Lease), tr.Header.Revision, ids)
	if loaded {
		r.barrier.Done()
		return group, true
	}

	return group, true
}

func (r *_Router) watchingForGroups() {
	defer r.barrier.Done()

	rsp, err := r.client.Get(r.ctx, r.groupIdKeyPrefix, etcdv3.WithPrefix(), etcdv3.WithKeysOnly())
	if err != nil {
		log.L(r.svcCtx).Panic("get groups keys failed", zap.Error(err))
		return
	}

	for _, kv := range rsp.Kvs {
		groupAddr, ok := r.parseGroupIdKey(string(kv.Key))
		if !ok {
			log.L(r.svcCtx).Warn("invalid group id key", zap.ByteString("key", kv.Key))
			continue
		}
		if _, ok := r.GetGroupByAddr(r.ctx, groupAddr); !ok {
			log.L(r.svcCtx).Warn("group not cached", zap.String("group_addr", groupAddr))
			continue
		}
	}

	revision := rsp.Header.Revision + 1

	log.L(r.svcCtx).Debug("watching for groups started", zap.String("key", r.groupIdKeyPrefix), zap.Int64("revision", revision))

	for watchRsp := range r.client.Watch(r.ctx, r.groupIdKeyPrefix, etcdv3.WithPrefix(), etcdv3.WithRev(revision)) {
		if watchRsp.Canceled {
			log.L(r.svcCtx).Debug("watching for groups canceled",
				zap.String("key", r.groupIdKeyPrefix),
				zap.Int64("revision", revision),
				zap.Error(watchRsp.Err()))
			break
		}
		if watchRsp.Err() != nil {
			log.L(r.svcCtx).Panic("watching for groups unexpectedly interrupted",
				zap.String("key", r.groupIdKeyPrefix),
				zap.Int64("revision", revision),
				zap.Error(watchRsp.Err()))
			break
		}

		for _, event := range watchRsp.Events {
			if event.Type != etcdv3.EventTypePut {
				continue
			}

			groupIdKey := string(event.Kv.Key)
			groupAddr, ok := r.parseGroupIdKey(groupIdKey)
			if !ok {
				log.L(r.svcCtx).Warn("invalid group id key", zap.String("key", groupIdKey))
				continue
			}

			if _, ok := r.GetGroupByAddr(r.ctx, groupAddr); !ok {
				log.L(r.svcCtx).Warn("group not cached", zap.String("group_addr", groupAddr))
			}
		}
	}

	log.L(r.svcCtx).Debug("watching for groups stopped", zap.String("key", r.groupIdKeyPrefix), zap.Int64("revision", revision))
}

func (r *_Router) cacheGroup(groupName, groupAddr string, leaseId etcdv3.LeaseID, revision int64, ids []uid.Id) (*_Group, bool) {
	if cached, ok := r.groups.Load(groupName); ok {
		exists := cached.(*_Group)
		if exists.leaseId == leaseId {
			return exists, true
		}
	}

	group := &_Group{}
	group.init(r, groupAddr, leaseId, revision, ids)

	cached, loaded := r.groups.LoadOrStore(groupName, group)
	if !loaded {
		log.L(r.svcCtx).Info("group cached",
			zap.String("group_name", group.Name()),
			zap.String("group_addr", group.ClientAddr()))
		r.groupCount.Add(1)
		go group.watchingForChanges()
		return group, false
	}

	exists := cached.(*_Group)
	if exists == group || exists.leaseId == leaseId {
		group.markExpired()
		return exists, true
	}

	if exists.latestRevision < revision && r.groups.CompareAndSwap(groupName, exists, group) {
		log.L(r.svcCtx).Info("group cache replaced",
			zap.String("group_name", groupName),
			zap.String("group_addr", group.ClientAddr()),
			zap.Int64("prev_lease_id", int64(exists.leaseId)),
			zap.Int64("curr_lease_id", int64(group.leaseId)))
		exists.markExpired()
		go group.watchingForChanges()
		return group, false
	}

	group.markExpired()
	if exists.latestRevision < revision {
		return r.cacheGroup(groupName, groupAddr, leaseId, revision, ids)
	}
	return exists, true
}

func (r *_Router) uncacheGroup(group *_Group) {
	if group == nil {
		return
	}
	if r.groups.CompareAndDelete(group.Name(), group) {
		log.L(r.svcCtx).Info("group uncached",
			zap.String("group_name", group.Name()),
			zap.String("group_addr", group.ClientAddr()))
		r.groupCount.Add(-1)
	}
}

func (r *_Router) revokeGroupLease(ctx context.Context, leaseID etcdv3.LeaseID) error {
	_, err := r.client.Revoke(ctx, leaseID)
	if err != nil && !errors.Is(err, rpctypes.ErrLeaseNotFound) {
		return err
	}
	return nil
}

func (r *_Router) groupIdKey(groupAddr string) string {
	return path.Join(r.groupIdKeyPrefix, groupAddr)
}

func (r *_Router) groupEntitiesPrefix(groupAddr string) string {
	return path.Join(r.groupEntitiesKeyPrefix, groupAddr) + "/"
}

func (r *_Router) entityGroupsPrefix(entityId uid.Id) string {
	return path.Join(r.entityGroupsKeyPrefix, entityId.String()) + "/"
}

func (r *_Router) parseGroupIdKey(key string) (string, bool) {
	groupAddr := strings.TrimPrefix(key, r.groupIdKeyPrefix)
	if groupAddr == key || groupAddr == "" {
		return "", false
	}
	return groupAddr, true
}

func (r *_Router) parseGroupEntitiesKey(key string) (groupAddr string, entityId uid.Id, ok bool) {
	trimmed := strings.TrimPrefix(key, r.groupEntitiesKeyPrefix)
	if trimmed == key {
		return "", uid.Nil, false
	}

	idx := strings.LastIndex(trimmed, "/")
	if idx <= 0 || idx >= len(trimmed)-1 {
		return "", uid.Nil, false
	}

	return trimmed[:idx], uid.From(trimmed[idx+1:]), true
}

func (r *_Router) parseEntityGroupsKey(key string) (entityId uid.Id, groupAddr string, ok bool) {
	trimmed := strings.TrimPrefix(key, r.entityGroupsKeyPrefix)
	if trimmed == key {
		return uid.Nil, "", false
	}

	idx := strings.Index(trimmed, "/")
	if idx <= 0 || idx >= len(trimmed)-1 {
		return uid.Nil, "", false
	}

	return uid.From(trimmed[:idx]), trimmed[idx+1:], true
}
