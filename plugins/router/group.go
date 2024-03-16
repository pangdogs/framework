package router

import (
	"context"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/util/generic"
	"git.golaxy.org/core/util/uid"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/transport"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/util/binaryutil"
	"github.com/elliotchance/pie/v2"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"path"
	"slices"
	"sync"
)

// IGroup 分组接口
type IGroup interface {
	context.Context
	// GetId 获取分组Id
	GetId() uid.Id
	// AddEntities 添加实体
	AddEntities(ctx context.Context, entIds ...uid.Id) error
	// RemoveEntities 删除实体
	RemoveEntities(ctx context.Context, entIds ...uid.Id) error
	// RenewEntities 刷新所有实体
	RenewEntities(ctx context.Context, entIds ...uid.Id) error
	// RangeEntities 遍历所有实体
	RangeEntities(fun generic.Func1[uid.Id, bool])
	// CountEntities 获取实体数量
	CountEntities() int
	// RefreshTTL 刷新TTL
	RefreshTTL(ctx context.Context) error
	// SendData 发送数据
	SendData(data []byte) error
	// SendEvent 发送自定义事件
	SendEvent(event transport.Event[gtp.MsgReader]) error
	// SendDataChan 发送数据的channel
	SendDataChan() chan<- binaryutil.RecycleBytes
	// SendEventChan 发送自定义事件的channel
	SendEventChan() chan<- transport.Event[gtp.MsgReader]
}

func (r *_Router) newGroup(groupId uid.Id, groupKey string, leaseId etcdv3.LeaseID, revision int64, entIds []uid.Id) *_Group {
	ctx, cancel := context.WithCancel(r.servCtx)

	return &_Group{
		Context:   ctx,
		terminate: cancel,
		router:    r,
		deleted:   false,
		id:        groupId,
		groupKey:  groupKey,
		leaseId:   leaseId,
		revision:  revision,
		entities:  entIds,
	}
}

func (r *_Router) addGroup(group *_Group) (*_Group, error) {
	var newer, ret *_Group
	var err error

	r.groups.AutoLock(func(groups *map[uid.Id]*_Group) {
		if exists, ok := (*groups)[group.id]; ok {
			exists.Lock()
			defer exists.Unlock()

			if exists.revision > group.revision {
				ret = nil
				err = ErrGroupExisted
				return
			} else if exists.revision == group.revision {
				ret = exists
				err = nil
				return
			}
		}

		(*groups)[group.id] = group
		newer = group
		ret = group
		err = nil
	})

	if newer != nil && r.options.GroupAutoRefreshTTL {
		rspChan, err := r.client.KeepAlive(newer.Context, newer.leaseId)
		if err != nil {
			log.Errorf(r.servCtx, "etcd keepalive %q failed, %s", newer.groupKey, err)
		} else {
			go func() {
				for range rspChan {
				}
			}()
		}
	}

	return ret, err
}

func (r *_Router) getOrAddGroup(group *_Group) (*_Group, error) {
	var newer, ret *_Group
	var err error

	r.groups.AutoLock(func(groups *map[uid.Id]*_Group) {
		if exists, ok := (*groups)[group.id]; ok {
			exists.Lock()
			defer exists.Unlock()

			if exists.revision >= group.revision {
				ret = exists
				err = nil
				return
			}
		}

		(*groups)[group.id] = group
		newer = group
		ret = group
		err = nil
	})

	if newer != nil && r.options.GroupAutoRefreshTTL {
		rspChan, err := r.client.KeepAlive(newer.Context, newer.leaseId)
		if err != nil {
			log.Errorf(r.servCtx, "etcd keepalive %q failed, %s", newer.groupKey, err)
		} else {
			go func() {
				for range rspChan {
				}
			}()
		}
	}

	return ret, err
}

type _Group struct {
	sync.RWMutex
	context.Context
	terminate context.CancelFunc
	router    *_Router
	deleted   bool
	id        uid.Id
	groupKey  string
	leaseId   etcdv3.LeaseID
	revision  int64
	entities  []uid.Id
}

// GetId 获取分组Id
func (g *_Group) GetId() uid.Id {
	return g.id
}

// AddEntities 添加实体
func (g *_Group) AddEntities(ctx context.Context, entIds ...uid.Id) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if len(entIds) <= 0 {
		return fmt.Errorf("%w: entIds is empty", core.ErrArgs)
	}

	if len(entIds) > 1 {
		entIds = pie.Unique(entIds)
	}

	opsPut := make([]etcdv3.Op, 0, len(entIds))
	for _, entId := range entIds {
		opsPut = append(opsPut, etcdv3.OpPut(path.Join(g.groupKey, entId.String()), "", etcdv3.WithLease(g.leaseId), etcdv3.WithIgnoreValue()))
	}

	rsp, err := g.router.client.Txn(ctx).
		Then(opsPut...).
		Commit()
	if err != nil {
		return err
	}

	g.Lock()
	defer g.Unlock()

	if g.revision < rsp.Header.Revision {
		g.revision = rsp.Header.Revision
	}

	g.entities = append(g.entities, entIds...)

	return nil
}

// RemoveEntities 删除实体
func (g *_Group) RemoveEntities(ctx context.Context, entIds ...uid.Id) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if len(entIds) <= 0 {
		return fmt.Errorf("%w: entIds is empty", core.ErrArgs)
	}

	if len(entIds) > 1 {
		entIds = pie.Unique(entIds)
	}

	opsDel := make([]etcdv3.Op, 0, len(entIds))
	for _, entId := range entIds {
		opsDel = append(opsDel, etcdv3.OpDelete(path.Join(g.groupKey, entId.String())))
	}

	rsp, err := g.router.client.Txn(ctx).
		Then(opsDel...).
		Commit()
	if err != nil {
		return err
	}

	g.Lock()
	defer g.Unlock()

	if g.revision < rsp.Header.Revision {
		g.revision = rsp.Header.Revision
	}

	g.entities = slices.DeleteFunc(g.entities, func(id uid.Id) bool {
		return pie.Contains(entIds, id)
	})

	return nil
}

// RenewEntities 刷新所有实体
func (g *_Group) RenewEntities(ctx context.Context, entIds ...uid.Id) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if len(entIds) > 1 {
		entIds = pie.Unique(entIds)
	}

	opDel := etcdv3.OpDelete(g.groupKey, etcdv3.WithPrefix())

	opsPut := make([]etcdv3.Op, 0, len(entIds)+1)
	opsPut = append(opsPut, etcdv3.OpPut(g.groupKey, "", etcdv3.WithLease(g.leaseId)))
	for _, entId := range entIds {
		opsPut = append(opsPut, etcdv3.OpPut(path.Join(g.groupKey, entId.String()), "", etcdv3.WithLease(g.leaseId)))
	}

	rsp, err := g.router.client.Txn(ctx).
		Then(opDel).
		Then(opsPut...).
		Commit()
	if err != nil {
		return err
	}

	g.Lock()
	defer g.Unlock()

	g.revision = rsp.Header.Revision
	g.entities = entIds

	return nil
}

// RangeEntities 遍历所有实体
func (g *_Group) RangeEntities(fun generic.Func1[uid.Id, bool]) {
	g.RLock()
	copied := slices.Clone(g.entities)
	g.RUnlock()

	for i := range copied {
		if !fun.Exec(copied[i]) {
			return
		}
	}
}

// CountEntities 获取实体数量
func (g *_Group) CountEntities() int {
	g.RLock()
	defer g.Unlock()
	return len(g.entities)
}

// RefreshTTL 刷新TTL
func (g *_Group) RefreshTTL(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	_, err := g.router.client.KeepAliveOnce(ctx, g.leaseId)
	return err
}

// SendData 发送数据
func (g *_Group) SendData(data []byte) error {
	g.RLock()
	defer g.RUnlock()

	if g.deleted {
		return ErrGroupDeleted
	}

	for i := range g.entities {
		entId := g.entities[i]

		g.router.servCtx.CallVoid(entId, func(entity ec.Entity, _ ...any) {
			session, ok := g.router.LookupSession(entity.GetId())
			if !ok {
				return
			}
			session.SendData(data)
		})
	}

	return nil
}

// SendEvent 发送自定义事件
func (g *_Group) SendEvent(event transport.Event[gtp.MsgReader]) error {
	return nil
}

// SendDataChan 发送数据的channel
func (g *_Group) SendDataChan() chan<- binaryutil.RecycleBytes {
	return nil
}

// SendEventChan 发送自定义事件的channel
func (g *_Group) SendEventChan() chan<- transport.Event[gtp.MsgReader] {
	return nil
}
