package framework

import (
	"context"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/pt"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/generic"
	"git.golaxy.org/framework/plugins/broker"
	"git.golaxy.org/framework/plugins/broker/nats_broker"
	"git.golaxy.org/framework/plugins/conf"
	"git.golaxy.org/framework/plugins/dentq"
	"git.golaxy.org/framework/plugins/discovery"
	"git.golaxy.org/framework/plugins/discovery/etcd_discovery"
	"git.golaxy.org/framework/plugins/dserv"
	"git.golaxy.org/framework/plugins/dsync"
	"git.golaxy.org/framework/plugins/dsync/etcd_dsync"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/plugins/log/zap_log"
	"git.golaxy.org/framework/plugins/rpc"
	"github.com/spf13/viper"
	etcd_client "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type _IService interface {
	init(startupConf *viper.Viper, name string, composite any)
	generate(ctx context.Context) core.Service
}

// ServiceBehavior 服务行为，在开发新服务时，匿名嵌入至服务结构体中
type ServiceBehavior struct {
	startupConf *viper.Viper
	name        string
	composite   any
}

func (sb *ServiceBehavior) init(startupConf *viper.Viper, name string, composite any) {
	sb.startupConf = startupConf
	sb.name = name
	sb.composite = composite
}

func (sb *ServiceBehavior) generate(ctx context.Context) core.Service {
	startupConf := sb.GetStartupConf()

	memKVs := &sync.Map{}
	memKVs.Store("startup.conf", startupConf)

	ctx = context.WithValue(ctx, "mem_kvs", memKVs)

	autoRecover := startupConf.GetBool("service.auto_recover")
	var reportError chan error

	if autoRecover {
		reportError = make(chan error, 128)
	}

	servCtx := service.NewContext(
		service.With.Context(ctx),
		service.With.Name(sb.GetName()),
		service.With.PanicHandling(autoRecover, reportError),
		service.With.EntityLib(pt.NewEntityLib(pt.DefaultComponentLib())),
		service.With.RunningHandler(generic.CastDelegateAction2(func(ctx service.Context, state service.RunningState) {
			// 状态变化回调
			switch state {
			case service.RunningState_Birth:
				if cb, ok := sb.composite.(LifecycleServiceBirth); ok {
					cb.Birth(ctx)
				}
			case service.RunningState_Starting:
				if cb, ok := sb.composite.(LifecycleServiceStarting); ok {
					cb.Starting(ctx)
				}
			case service.RunningState_Started:
				if cb, ok := sb.composite.(LifecycleServiceStarted); ok {
					cb.Started(ctx)
				}
			case service.RunningState_Terminating:
				if cb, ok := sb.composite.(LifecycleServiceTerminating); ok {
					cb.Terminating(ctx)
				}
			case service.RunningState_Terminated:
				if cb, ok := sb.composite.(LifecycleServiceTerminated); ok {
					cb.Terminated(ctx)
				}

				if v, ok := memKVs.Load("zap.logger"); ok {
					v.(*zap.Logger).Sync()
				}

				if v, ok := memKVs.Load("etcd.client"); ok {
					v.(*etcd_client.Client).Close()
				}
			}
		})),
	)

	// 安装日志插件
	if cb, ok := sb.composite.(InstallServiceLogger); ok {
		cb.InstallLogger(servCtx)
	}
	if _, ok := servCtx.GetPluginBundle().Get(log.Name); !ok {
		level, err := zapcore.ParseLevel(startupConf.GetString("log.level"))
		if err != nil {
			panic(fmt.Errorf("parse startup config log.level failed, %s", err))
		}

		zapLogger, zapAtomicLevel := zap_log.NewJsonZapLogger(
			level,
			filepath.Join(startupConf.GetString("log.dir"), fmt.Sprintf("%s.%s.log", strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0])), sb.GetName())),
			startupConf.GetInt("log.size"),
			startupConf.GetBool("log.stdout"),
			startupConf.GetBool("log.development"),
		)

		memKVs.Store("zap.logger", zapLogger)
		memKVs.Store("zap.atomic_level", zapAtomicLevel)

		zap_log.Install(servCtx,
			zap_log.With.ZapLogger(zapLogger),
			zap_log.With.ServiceInfo(true),
		)
	}

	// 安装配置插件
	if cb, ok := sb.composite.(InstallServiceConfig); ok {
		cb.InstallConfig(servCtx)
	}
	if _, ok := servCtx.GetPluginBundle().Get(conf.Name); !ok {
		conf.Install(servCtx,
			conf.With.Format(startupConf.GetString("conf.format")),
			conf.With.LocalPath(startupConf.GetString("conf.local_path")),
			conf.With.Remote(
				startupConf.GetString("conf.remote_provider"),
				startupConf.GetString("conf.remote_endpoint"),
				startupConf.GetString("conf.remote_path"),
			),
			conf.With.AutoUpdate(startupConf.GetBool("conf.auto_update")),
		)
	}

	// 安装broker插件
	if cb, ok := sb.composite.(InstallServiceBroker); ok {
		cb.InstallBroker(servCtx)
	}
	if _, ok := servCtx.GetPluginBundle().Get(broker.Name); !ok {
		nats_broker.Install(servCtx,
			nats_broker.With.CustomAddresses(startupConf.GetString("nats.address")),
			nats_broker.With.CustomAuth(
				startupConf.GetString("nats.username"),
				startupConf.GetString("nats.password"),
			),
		)
	}

	// 安装服务发现插件
	if cb, ok := sb.composite.(InstallServiceRegistry); ok {
		cb.InstallRegistry(servCtx)
	}
	if _, ok := servCtx.GetPluginBundle().Get(discovery.Name); !ok {
		etcd_discovery.Install(servCtx,
			etcd_discovery.With.CustomAddresses(startupConf.GetString("etcd.address")),
			etcd_discovery.With.CustomAuth(
				startupConf.GetString("etcd.username"),
				startupConf.GetString("etcd.password"),
			),
		)
	}

	// 安装分布式同步插件
	if cb, ok := sb.composite.(InstallServiceDistSync); ok {
		cb.InstallDistSync(servCtx)
	}
	if _, ok := servCtx.GetPluginBundle().Get(dsync.Name); !ok {
		etcd_dsync.Install(servCtx,
			etcd_dsync.With.CustomAddresses(startupConf.GetString("etcd.address")),
			etcd_dsync.With.CustomAuth(
				startupConf.GetString("etcd.username"),
				startupConf.GetString("etcd.password"),
			),
		)
	}

	// 安装分布式服务插件
	if cb, ok := sb.composite.(InstallServiceDistService); ok {
		cb.InstallDistService(servCtx)
	}
	if _, ok := servCtx.GetPluginBundle().Get(dserv.Name); !ok {
		dserv.Install(servCtx,
			dserv.With.Version(startupConf.GetString(fmt.Sprintf("%s.version", sb.GetName()))),
			dserv.With.Meta(startupConf.GetStringMapString(fmt.Sprintf("%s.meta", sb.GetName()))),
			dserv.With.TTL(startupConf.GetDuration("service.ttl")),
			dserv.With.FutureTimeout(startupConf.GetDuration("service.future_timeout")),
		)
	}

	// 安装RPC支持插件
	if cb, ok := sb.composite.(InstallServiceRPC); ok {
		cb.InstallRPC(servCtx)
	}
	if _, ok := servCtx.GetPluginBundle().Get(rpc.Name); !ok {
		rpc.Install(servCtx)
	}

	// 安装分布式实体查询插件
	if cb, ok := sb.composite.(InstallServiceDistEntityQuerier); ok {
		cb.InstallDistEntityQuerier(servCtx)
	}
	if _, ok := servCtx.GetPluginBundle().Get(dentq.Name); !ok {
		dentq.Install(servCtx,
			dentq.With.CustomAddresses(startupConf.GetString("etcd.address")),
			dentq.With.CustomAuth(
				startupConf.GetString("etcd.username"),
				startupConf.GetString("etcd.password"),
			),
		)
	}

	// etcd连接初始化函数
	memKVs.Store("etcd.init_client", sync.OnceFunc(func() {
		cli, err := etcd_client.New(etcd_client.Config{
			Endpoints: []string{startupConf.GetString("etcd.address")},
			Username:  startupConf.GetString("etcd.username"),
			Password:  startupConf.GetString("etcd.password"),
		})
		if err != nil {
			panic(fmt.Errorf("new etcd client failed, %s", err))
		}
		memKVs.Store("etcd.client", cli)
	}))

	// 初始化回调
	if cb, ok := sb.composite.(LifecycleServiceInit); ok {
		cb.Init(servCtx)
	}

	// 自动恢复时，打印panic信息
	if servCtx.GetAutoRecover() && servCtx.GetReportError() != nil {
		go func() {
			for {
				select {
				case err := <-servCtx.GetReportError():
					log.Errorf(servCtx, "recover:\n%s", err)
				case <-servCtx.Done():
					return
				}
			}
		}()
	}

	return core.NewService(servCtx)
}

// GetName 获取服务名称
func (sb *ServiceBehavior) GetName() string {
	return sb.name
}

// GetStartupConf 获取启动参数配置
func (sb *ServiceBehavior) GetStartupConf() *viper.Viper {
	return sb.startupConf
}
