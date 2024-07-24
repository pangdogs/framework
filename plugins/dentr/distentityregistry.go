package dentr

import (
	"context"
	"crypto/tls"
	"errors"
	"git.golaxy.org/core"
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/event"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/plugins/log"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"math"
	"path"
	"time"
)

// IDistEntityRegistry 分布式实体注册支持，会将全局可以访问的实体注册为分布式实体
type IDistEntityRegistry interface {
	IDistEntityRegistryEventTab
}

func newDistEntityRegistry(settings ...option.Setting[DistEntityRegistryOptions]) IDistEntityRegistry {
	return &_DistEntityRegistry{
		options: option.Make(With.Default(), settings...),
	}
}

type _DistEntityRegistry struct {
	distEntityRegistryEventTab
	rtCtx   runtime.Context
	options DistEntityRegistryOptions
	client  *etcdv3.Client
	leaseId etcdv3.LeaseID
}

// InitRP 初始化运行时插件
func (d *_DistEntityRegistry) InitRP(ctx runtime.Context) {
	log.Debugf(ctx, "init plugin %q", self.Name)

	d.rtCtx = ctx
	d.rtCtx.ActivateEvent(&d.distEntityRegistryEventTab, event.EventRecursion_Allow)

	if d.options.EtcdClient == nil {
		cli, err := etcdv3.New(d.configure())
		if err != nil {
			log.Panicf(d.rtCtx, "new etcd client failed, %s", err)
		}
		d.client = cli
	} else {
		d.client = d.options.EtcdClient
	}

	for _, ep := range d.client.Endpoints() {
		func() {
			ctx, cancel := context.WithTimeout(d.rtCtx, 3*time.Second)
			defer cancel()

			if _, err := d.client.Status(ctx, ep); err != nil {
				log.Panicf(d.rtCtx, "status etcd %q failed, %s", ep, err)
			}
		}()
	}

	// 申请租约
	leaseId, err := d.grantLease()
	if err != nil {
		log.Panicf(d.rtCtx, "grant lease failed, %s", err)
	}
	d.leaseId = leaseId
	log.Debugf(d.rtCtx, "grant lease %d", d.leaseId)

	// 刷新实体信息
	d.rtCtx.GetEntityMgr().RangeEntities(d.register)

	// 租约心跳
	core.Await(d.rtCtx, core.TimeTick(d.rtCtx, d.options.TTL/2)).Pipe(nil, d.keepAliveLease)

	// 绑定事件
	d.rtCtx.ManagedHooks(
		runtime.BindEventEntityMgrAddEntity(ctx.GetEntityMgr(), d, 1000),
		runtime.BindEventEntityMgrRemoveEntity(ctx.GetEntityMgr(), d, -1000),
	)
}

// ShutRP 关闭运行时插件
func (d *_DistEntityRegistry) ShutRP(ctx runtime.Context) {
	log.Debugf(ctx, "shut plugin %q", self.Name)

	// 废除租约
	_, err := d.client.Revoke(context.Background(), d.leaseId)
	if err != nil {
		log.Errorf(d.rtCtx, "revoke lease %d failed, %s", d.leaseId, err)
	}

	if d.options.EtcdClient == nil {
		if d.client != nil {
			d.client.Close()
		}
	}

	d.distEntityRegistryEventTab.Close()
}

// OnEntityMgrAddEntity 实体管理器添加实体
func (d *_DistEntityRegistry) OnEntityMgrAddEntity(entityMgr runtime.EntityMgr, entity ec.Entity) {
	d.register(entity)
}

// OnEntityMgrRemoveEntity 实体管理器删除实体
func (d *_DistEntityRegistry) OnEntityMgrRemoveEntity(entityMgr runtime.EntityMgr, entity ec.Entity) {
	d.deregister(entity)
}

func (d *_DistEntityRegistry) register(entity ec.Entity) bool {
	if entity.GetScope() != ec.Scope_Global {
		return true
	}

	key := d.getEntityPath(entity)

	_, err := d.client.Put(d.rtCtx, key, "", etcdv3.WithLease(d.leaseId))
	if err != nil {
		log.Errorf(d.rtCtx, "put %q with lease %d failed, %s", key, d.leaseId, err)
		return false
	}
	log.Debugf(d.rtCtx, "put %q with lease %d ok", key, d.leaseId)

	// 通知分布式实体上线
	_EmitEventDistEntityOnline(d, entity)
	return true
}

func (d *_DistEntityRegistry) deregister(entity ec.Entity) {
	if entity.GetScope() != ec.Scope_Global {
		return
	}

	select {
	case <-d.rtCtx.Done():
		break
	default:
		key := d.getEntityPath(entity)

		_, err := d.client.Delete(d.rtCtx, key)
		if err != nil {
			log.Warnf(d.rtCtx, "delete %q failed, %s", key, err)
		} else {
			log.Debugf(d.rtCtx, "delete %q ok", key)
		}
	}

	// 通知分布式实体下线
	_EmitEventDistEntityOffline(d, entity)
}

func (d *_DistEntityRegistry) getEntityPath(entity ec.Entity) string {
	servCtx := service.Current(d.rtCtx)
	return path.Join(d.options.KeyPrefix, entity.GetId().String(), servCtx.GetName(), servCtx.GetId().String())
}

func (d *_DistEntityRegistry) keepAliveLease(ctx runtime.Context, ret async.Ret, args ...any) {
	// 刷新租约
	_, err := d.client.KeepAliveOnce(d.rtCtx, d.leaseId)
	if err == nil {
		log.Debugf(d.rtCtx, "keep alive lease %d ok", d.leaseId)
		return
	}

	if !errors.Is(err, rpctypes.ErrLeaseNotFound) {
		log.Errorf(d.rtCtx, "keep alive lease %d failed, %s", d.leaseId, err)
		return
	}

	// 通知所有分布式实体下线
	d.rtCtx.GetEntityMgr().RangeEntities(func(entity ec.Entity) bool {
		if entity.GetScope() == ec.Scope_Global {
			_EmitEventDistEntityOffline(d, entity)
		}
		return true
	})

	log.Debugf(d.rtCtx, "lease %d not found, try grant a new lease", d.leaseId)

	// 重新申请租约
	leaseId, err := d.grantLease()
	if err != nil {
		log.Errorf(d.rtCtx, "grant new lease failed, %s", err)
		return
	}
	d.leaseId = leaseId
	log.Debugf(d.rtCtx, "grant new lease %d", d.leaseId)

	// 刷新实体信息
	d.rtCtx.GetEntityMgr().RangeEntities(d.register)
}

func (d *_DistEntityRegistry) grantLease() (etcdv3.LeaseID, error) {
	lgr, err := d.client.Grant(d.rtCtx, int64(math.Ceil(d.options.TTL.Seconds())))
	if err != nil {
		return etcdv3.NoLease, err
	}
	return lgr.ID, nil
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
