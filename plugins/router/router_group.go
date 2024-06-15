package router

import (
	"context"
	"errors"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/net/gtp/transport"
	"git.golaxy.org/framework/plugins/gate"
	"git.golaxy.org/framework/util/binaryutil"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"math"
	"path"
	"strconv"
)

// AddGroup 添加分组
func (r *_Router) AddGroup(ctx context.Context, groupAddr string) (IGroup, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if !gate.CliDetails.InMulticastSubdomain(groupAddr) {
		return nil, fmt.Errorf("%w: incorrect groupAddr", core.ErrArgs)
	}

	lgr, err := r.client.Grant(ctx, int64(math.Ceil(r.options.GroupTTL.Seconds())))
	if err != nil {
		return nil, err
	}
	leaseId := lgr.ID

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

	cached := r.groupCache.Set(groupAddr, group, tr.Header.Revision, 0)
	if cached == group {
		go group.mainLoop()
	}

	return cached, nil
}

// DeleteGroup 删除分组
func (r *_Router) DeleteGroup(ctx context.Context, groupAddr string) {
	if ctx == nil {
		ctx = context.Background()
	}

	if !gate.CliDetails.InMulticastSubdomain(groupAddr) {
		return
	}

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
func (r *_Router) GetGroup(ctx context.Context, groupAddr string) (IGroup, bool) {
	if ctx == nil {
		ctx = context.Background()
	}

	if !gate.CliDetails.InMulticastSubdomain(groupAddr) {
		return nil, false
	}

	group, ok := r.groupCache.Get(groupAddr)
	if ok {
		return group, true
	}

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

	group = r.newGroup(groupAddr, leaseId, tr.Header.Revision, entIds)

	cached := r.groupCache.Set(groupAddr, group, tr.Header.Revision, 0)
	if cached == group {
		go group.mainLoop()
	}

	return cached, true
}

// RangeGroups 遍历实体所在的分组
func (r *_Router) RangeGroups(ctx context.Context, entityId uid.Id, fun generic.Func1[IGroup, bool]) {
	if ctx == nil {
		ctx = context.Background()
	}

	if _, ok := r.servCtx.GetEntityMgr().GetEntity(entityId); !ok {
		return
	}

	groupAddrs, ok := r.entityGroupsCache.Get(entityId)
	if !ok {
		gr, err := r.client.Get(ctx, path.Join(r.options.EntityGroupsKeyPrefix, entityId.String())+"/",
			etcdv3.WithPrefix(),
			etcdv3.WithSort(etcdv3.SortByModRevision, etcdv3.SortDescend),
			etcdv3.WithIgnoreValue())
		if err != nil {
			return
		}

		if len(gr.Kvs) <= 0 {
			return
		}

		groupAddrs = make([]string, 0, len(gr.Kvs))

		for _, kv := range gr.Kvs {
			groupAddrs = append(groupAddrs, string(kv.Key))
		}

		groupAddrs = r.entityGroupsCache.Set(entityId, groupAddrs, gr.Header.Revision, r.options.EntityGroupsCacheTTL)
	}

	for _, groupAddr := range groupAddrs {
		if group, ok := r.GetGroup(ctx, groupAddr); ok {
			if !fun.Exec(group) {
				return
			}
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
