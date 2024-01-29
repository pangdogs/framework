package dent

import (
	"context"
	"crypto/tls"
	"errors"
	"git.golaxy.org/core"
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/event"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/framework/plugins/log"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	etcd_client "go.etcd.io/etcd/client/v3"
	"math"
	"path"
	"time"
)

// IDistEntities 分布式实体支持，会将全局可以访问的实体注册为分布式实体
type IDistEntities interface {
	IDistEntitiesEventTab
}

func newDistEntities(settings ...option.Setting[DistEntitiesOptions]) IDistEntities {
	return &_DistEntities{
		options: option.Make(Option{}.Default(), settings...),
	}
}

type _DistEntities struct {
	distEntitiesEventTab
	options DistEntitiesOptions
	rtCtx   runtime.Context
	client  *etcd_client.Client
	leaseId etcd_client.LeaseID
}

// InitRP 初始化运行时插件
func (d *_DistEntities) InitRP(ctx runtime.Context) {
	log.Infof(ctx, "init plugin %q", plugin.Name)

	d.rtCtx = ctx
	d.rtCtx.ActivateEvent(&d.distEntitiesEventTab, event.EventRecursion_Allow)

	if d.options.EtcdClient == nil {
		cli, err := etcd_client.New(d.configure())
		if err != nil {
			log.Panicf(d.rtCtx, "new etcd client failed, %s", err)
		}
		d.client = cli
	} else {
		d.client = d.options.EtcdClient
	}

	for _, ep := range d.client.Endpoints() {
		if _, err := d.client.Status(d.rtCtx, ep); err != nil {
			log.Panicf(d.rtCtx, "status etcd %q failed, %s", ep, err)
		}
	}

	// 申请租约
	if err := d.grantLease(); err != nil {
		log.Panicf(d.rtCtx, "grant lease failed, %s", err)
	}
	log.Debugf(d.rtCtx, "grant lease %d", d.leaseId)

	// 刷新实体信息
	d.rtCtx.GetEntityMgr().RangeEntities(d.register)

	// 租约心跳
	core.Await(d.rtCtx, core.TimeTick(d.rtCtx, d.options.TTL/2)).Pipe(d.rtCtx, d.keepAliveLease)

	// 绑定事件
	d.rtCtx.ManagedHooks(
		runtime.BindEventEntityMgrAddEntity(ctx.GetEntityMgr(), d, 1000),
		runtime.BindEventEntityMgrRemovingEntity(ctx.GetEntityMgr(), d, -1000),
	)
}

// ShutRP 关闭运行时插件
func (d *_DistEntities) ShutRP(ctx runtime.Context) {
	log.Infof(ctx, "shut plugin %q", plugin.Name)

	// 废除租约
	_, err := d.client.Revoke(context.Background(), d.leaseId)
	if err != nil {
		log.Errorf(d.rtCtx, "revoke lease %d failed, %s", d.leaseId, err)
	}

	d.distEntitiesEventTab.Close()
}

// OnEntityMgrAddEntity 实体管理器添加实体
func (d *_DistEntities) OnEntityMgrAddEntity(entityMgr runtime.EntityMgr, entity ec.Entity) {
	d.register(entity)
}

// OnEntityMgrRemovingEntity 实体管理器开始删除实体
func (d *_DistEntities) OnEntityMgrRemovingEntity(entityMgr runtime.EntityMgr, entity ec.Entity) {
	select {
	case <-d.rtCtx.Done():
		return
	default:
	}
	d.deregister(entity)
}

func (d *_DistEntities) register(entity ec.Entity) bool {
	if entity.GetScope() != ec.Scope_Global {
		return true
	}

	key := d.getEntityPath(entity)

	_, err := d.client.Put(d.rtCtx, key, "", etcd_client.WithLease(d.leaseId))
	if err != nil {
		log.Errorf(d.rtCtx, "put %q with lease %d failed, %s", key, d.leaseId, err)
		return false
	}

	log.Debugf(d.rtCtx, "put %q with lease %d ok", key, d.leaseId)

	// 通知分布式实体上线
	emitEventDistEntityOnline(d, entity)
	return true
}

func (d *_DistEntities) deregister(entity ec.Entity) {
	if entity.GetScope() != ec.Scope_Global {
		return
	}

	key := d.getEntityPath(entity)

	_, err := d.client.Delete(d.rtCtx, key)
	if err != nil {
		log.Errorf(d.rtCtx, "delete %q failed, %s", key, err)
		return
	}

	log.Debugf(d.rtCtx, "delete %q ok", key)

	// 通知分布式实体下线
	emitEventDistEntityOffline(d, entity)
}

func (d *_DistEntities) getEntityPath(entity ec.Entity) string {
	servCtx := service.Current(d.rtCtx)
	return path.Join(d.options.KeyPrefix, entity.GetId().String(), servCtx.GetName(), servCtx.GetId().String())
}

func (d *_DistEntities) keepAliveLease(ctx runtime.Context, ret runtime.Ret, args ...any) {
	// 刷新租约
	_, err := d.client.KeepAliveOnce(d.rtCtx, d.leaseId)
	if err == nil {
		log.Debugf(d.rtCtx, "keep alive lease %q ok", d.leaseId)
		return
	}

	if !errors.Is(err, rpctypes.ErrLeaseNotFound) {
		log.Errorf(d.rtCtx, "keep alive lease %q failed, %s", d.leaseId, err)
		return
	}

	// 通知所有分布式实体下线
	d.rtCtx.GetEntityMgr().RangeEntities(func(entity ec.Entity) bool {
		if entity.GetScope() == ec.Scope_Global {
			emitEventDistEntityOffline(d, entity)
		}
		return true
	})

	log.Warnf(d.rtCtx, "lease %d not found, try grant a new lease", d.leaseId)

	// 重新申请租约
	if err := d.grantLease(); err != nil {
		log.Errorf(d.rtCtx, "grant new lease failed, %s", err)
		return
	}
	log.Debugf(d.rtCtx, "grant new lease %d", d.leaseId)

	// 刷新实体信息
	d.rtCtx.GetEntityMgr().RangeEntities(d.register)
}

func (d *_DistEntities) grantLease() error {
	// 申请租约
	lgr, err := d.client.Grant(d.rtCtx, int64(math.Ceil(d.options.TTL.Seconds())))
	if err != nil {
		return err
	}
	d.leaseId = lgr.ID

	return nil
}

func (d *_DistEntities) configure() etcd_client.Config {
	if d.options.EtcdConfig != nil {
		return *d.options.EtcdConfig
	}

	config := etcd_client.Config{
		Endpoints:   d.options.CustomAddresses,
		Username:    d.options.CustomUsername,
		Password:    d.options.CustomPassword,
		DialTimeout: 3 * time.Second,
	}

	if d.options.CustomSecure || d.options.CustomTLSConfig != nil {
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
