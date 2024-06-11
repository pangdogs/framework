package router

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/net/netpath"
	"git.golaxy.org/framework/plugins/gate"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/util/concurrent"
	"github.com/elliotchance/pie/v2"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"maps"
	"math"
	"path"
	"strings"
	"time"
)

var (
	ErrEntityNotFound  = errors.New("router: entity not found")
	ErrSessionNotFound = errors.New("router: session not found")
	ErrEntityMapped    = errors.New("router: entity is already mapping")
	ErrSessionMapped   = errors.New("router: session is already mapping")
	ErrGroupExisted    = errors.New("router: group already existed")
	ErrGroupDeleted    = errors.New("router: group already deleted")
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
	// GetGroup 查询分组
	GetGroup(groupId uid.Id) (IGroup, bool)
	// GetOrAddGroup 查询或添加分组
	GetOrAddGroup(ctx context.Context, groupId uid.Id, ttl time.Duration, entIds ...uid.Id) (IGroup, error)
	// GetAndDeleteGroup 查询并删除分组
	GetAndDeleteGroup(ctx context.Context, groupId uid.Id) (IGroup, bool)
	// AddGroup 添加分组
	AddGroup(ctx context.Context, groupId uid.Id, ttl time.Duration, entIds ...uid.Id) (IGroup, error)
	// DeleteGroup 删除分组
	DeleteGroup(ctx context.Context, groupId uid.Id)
	// RangeGroups 遍历所有分组
	RangeGroups(fun generic.Func1[IGroup, bool])
	// CountGroups 统计所有分组数量
	CountGroups() int
}

func newRouter(settings ...option.Setting[RouterOptions]) IRouter {
	return &_Router{
		options:  option.Make(With.Default(), settings...),
		planning: concurrent.MakeLockedMap[uid.Id, *_Mapping](0),
		groups:   concurrent.MakeLockedMap[uid.Id, *_Group](0),
	}
}

type _Router struct {
	options  RouterOptions
	servCtx  service.Context
	gate     gate.IGate
	client   *etcdv3.Client
	planning concurrent.LockedMap[uid.Id, *_Mapping]
	groups   concurrent.LockedMap[uid.Id, *_Group]
}

// InitSP 初始化服务插件
func (r *_Router) InitSP(ctx service.Context) {
	log.Infof(ctx, "init plugin %q", self.Name)

	r.servCtx = ctx
	r.gate = gate.Using(r.servCtx)

	if r.options.EtcdClient == nil {
		cli, err := etcdv3.New(r.configure())
		if err != nil {
			log.Panicf(ctx, "new etcd client failed, %s", err)
		}
		r.client = cli
	} else {
		r.client = r.options.EtcdClient
	}

	for _, ep := range r.client.Endpoints() {
		func() {
			ctx, cancel := context.WithTimeout(r.servCtx, 3*time.Second)
			defer cancel()

			if _, err := r.client.Status(ctx, ep); err != nil {
				log.Panicf(r.servCtx, "status etcd %q failed, %s", ep, err)
			}
		}()
	}

	go r.mainLoop()
}

// ShutSP 关闭服务插件
func (r *_Router) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut plugin %q", self.Name)

	if r.options.EtcdClient == nil {
		if r.client != nil {
			r.client.Close()
		}
	}
}

// Mapping 路由映射
func (r *_Router) Mapping(entityId, sessionId uid.Id) (IMapping, error) {
	entity, ok := r.servCtx.GetEntityMgr().GetEntity(entityId)
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

		ctx, cancel := context.WithCancel(r.servCtx)

		mapping := &_Mapping{
			Context:   ctx,
			router:    r,
			terminate: cancel,
			entity:    entity,
			session:   session,
			cliAddr:   netpath.Path(gate.CliDetails.PathSeparator, gate.CliDetails.NodeSubdomain, entity.GetId().String()),
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

// GetGroup 查询分组
func (r *_Router) GetGroup(groupId uid.Id) (IGroup, bool) {
	return r.groups.Get(groupId)
}

// GetOrAddGroup 查询或添加分组
func (r *_Router) GetOrAddGroup(ctx context.Context, groupId uid.Id, ttl time.Duration, entIds ...uid.Id) (IGroup, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if groupId == uid.Nil {
		return nil, fmt.Errorf("%w: groupId is nil", core.ErrArgs)
	}

	if ttl <= 0 {
		ttl = r.options.GroupTTL
	}

	if len(entIds) > 1 {
		entIds = pie.Unique(entIds)
	}

	lgr, err := r.client.Grant(ctx, int64(math.Ceil(ttl.Seconds())))
	if err != nil {
		return nil, err
	}
	leaseId := lgr.ID

	groupKey := path.Join(r.options.KeyPrefix, groupId.String())

	opsPut := make([]etcdv3.Op, 0, len(entIds)+1)
	opsPut = append(opsPut, etcdv3.OpPut(groupKey, "", etcdv3.WithLease(leaseId)))
	for _, entId := range entIds {
		opsPut = append(opsPut, etcdv3.OpPut(path.Join(groupKey, entId.String()), "", etcdv3.WithLease(leaseId)))
	}

	opGet := etcdv3.OpGet(groupKey,
		etcdv3.WithPrefix(),
		etcdv3.WithSort(etcdv3.SortByModRevision, etcdv3.SortAscend),
		etcdv3.WithIgnoreValue(),
		etcdv3.WithSerializable())

	rsp, err := r.client.Txn(ctx).
		If(etcdv3.Compare(etcdv3.Version(groupKey), "=", 0)).
		Then(opsPut...).
		Else(opGet).
		Commit()
	if err != nil {
		return nil, err
	}

	if rsp.Succeeded {
		return r.getOrAddGroup(r.newGroup(groupId, groupKey, leaseId, rsp.Header.Revision, entIds))
	} else {
		leaseId := etcdv3.NoLease
		entIds := make([]uid.Id, 0, len(rsp.Responses)-1)

		for _, kv := range rsp.Responses[0].GetResponseRange().Kvs {
			leaseId = etcdv3.LeaseID(kv.Lease)

			entId := strings.TrimPrefix(strings.TrimPrefix(string(kv.Key), groupKey), "/")
			if entId == "" {
				continue
			}

			entIds = append(entIds, uid.From(entId))
		}

		return r.getOrAddGroup(r.newGroup(groupId, groupKey, leaseId, rsp.Header.Revision, entIds))
	}
}

// GetAndDeleteGroup 查询并删除分组
func (r *_Router) GetAndDeleteGroup(ctx context.Context, groupId uid.Id) (IGroup, bool) {
	if ctx == nil {
		ctx = context.Background()
	}

	var group *_Group

	r.groups.AutoLock(func(groups *map[uid.Id]*_Group) {
		if group, _ = (*groups)[groupId]; group != nil {
			group.Lock()
			defer group.Lock()
			group.deleted = true
			delete(*groups, groupId)
		}
	})

	if group != nil {
		r.client.Revoke(ctx, group.leaseId)
		group.terminate()
		return group, true
	}

	return nil, false
}

// AddGroup 添加分组
func (r *_Router) AddGroup(ctx context.Context, groupId uid.Id, ttl time.Duration, entIds ...uid.Id) (IGroup, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if groupId == uid.Nil {
		return nil, fmt.Errorf("%w: groupId is nil", core.ErrArgs)
	}

	if ttl <= 0 {
		ttl = r.options.GroupTTL
	}

	if len(entIds) > 1 {
		entIds = pie.Unique(entIds)
	}

	lgr, err := r.client.Grant(ctx, int64(math.Ceil(ttl.Seconds())))
	if err != nil {
		return nil, err
	}
	leaseId := lgr.ID

	groupKey := path.Join(r.options.KeyPrefix, groupId.String())

	opsPut := make([]etcdv3.Op, 0, len(entIds)+1)
	opsPut = append(opsPut, etcdv3.OpPut(groupKey, "", etcdv3.WithLease(leaseId)))
	for _, entId := range entIds {
		opsPut = append(opsPut, etcdv3.OpPut(path.Join(groupKey, entId.String()), "", etcdv3.WithLease(leaseId)))
	}

	rsp, err := r.client.Txn(ctx).
		If(etcdv3.Compare(etcdv3.Version(groupKey), "=", 0)).
		Then(opsPut...).
		Commit()
	if err != nil {
		return nil, err
	}

	if !rsp.Succeeded {
		return nil, ErrGroupExisted
	}

	return r.addGroup(r.newGroup(groupId, groupKey, leaseId, rsp.Header.Revision, entIds))
}

// DeleteGroup 删除分组
func (r *_Router) DeleteGroup(ctx context.Context, groupId uid.Id) {
	if ctx == nil {
		ctx = context.Background()
	}

	var group *_Group

	r.groups.AutoLock(func(groups *map[uid.Id]*_Group) {
		if group, _ = (*groups)[groupId]; group != nil {
			group.Lock()
			defer group.Lock()
			group.deleted = true
			delete(*groups, groupId)
		}
	})

	if group != nil {
		r.client.Revoke(ctx, group.leaseId)
		group.terminate()
	}
}

// RangeGroups 遍历所有分组
func (r *_Router) RangeGroups(fun generic.Func1[IGroup, bool]) {
	var copied map[uid.Id]*_Group

	r.groups.AutoRLock(func(groups *map[uid.Id]*_Group) {
		copied = maps.Clone(*groups)
	})

	for _, v := range copied {
		if !fun.Exec(v) {
			return
		}
	}
}

// CountGroups 统计所有分组数量
func (r *_Router) CountGroups() int {
	return r.groups.Len()
}

func (r *_Router) mainLoop() {

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
