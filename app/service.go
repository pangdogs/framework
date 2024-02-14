package app

import (
	"fmt"
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
	"go.uber.org/zap/zapcore"
)

type _IService interface {
	init(app *App, name string, composite any)
	generate() service.Context
}

type ServiceBehavior struct {
	app       *App
	name      string
	composite any
}

func (sb *ServiceBehavior) init(app *App, name string, composite any) {
	sb.app = app
	sb.name = name
	sb.composite = composite
}

func (sb *ServiceBehavior) generate() service.Context {
	servCtx := service.NewContext(
		service.Option{}.Name(sb.GetName()),
		service.Option{}.EntityLib(pt.NewEntityLib(pt.DefaultComponentLib())),
		service.Option{}.RunningHandler(generic.CastDelegateAction2(func(ctx service.Context, state service.RunningState) {
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
			}
		})),
	)

	startupConf := sb.GetStartupConf()

	// 安装日志插件
	if cb, ok := sb.composite.(InstallLogger); ok {
		cb.InstallLogger(servCtx)
	}
	if _, ok := servCtx.GetPluginBundle().Get(log.Name); !ok {
		level, err := zapcore.ParseLevel(startupConf.GetString("log.level"))
		if err != nil {
			panic(fmt.Errorf("parse log.level failed, %s", err))
		}

		zapLogger, _ := zap_log.NewJsonZapLogger(level,
			startupConf.GetString("log.file"),
			startupConf.GetInt("log.size"),
			startupConf.GetBool("log.stdout"),
			startupConf.GetBool("log.development"),
		)

		zap_log.Install(servCtx,
			zap_log.Option{}.ZapLogger(zapLogger),
			zap_log.Option{}.ServiceInfo(true),
		)
	}

	// 安装配置插件
	if cb, ok := sb.composite.(InstallConfig); ok {
		cb.InstallConfig(servCtx)
	}
	if _, ok := servCtx.GetPluginBundle().Get(conf.Name); !ok {
		conf.Install(servCtx,
			conf.Option{}.Format(startupConf.GetString("conf.format")),
			conf.Option{}.LocalPath(startupConf.GetString("conf.local_path")),
			conf.Option{}.Remote(
				startupConf.GetString("conf.remote_provider"),
				startupConf.GetString("conf.remote_endpoint"),
				startupConf.GetString("conf.remote_path"),
			),
			conf.Option{}.AutoUpdate(startupConf.GetBool("conf.auto_update")),
		)
	}

	// 安装分布式broker插件
	if cb, ok := sb.composite.(InstallBroker); ok {
		cb.InstallBroker(servCtx)
	}
	if _, ok := servCtx.GetPluginBundle().Get(broker.Name); !ok {
		nats_broker.Install(servCtx,
			nats_broker.Option{}.CustomAddresses(startupConf.GetString("nats.address")),
			nats_broker.Option{}.CustomAuth(
				startupConf.GetString("nats.username"),
				startupConf.GetString("nats.password"),
			),
		)
	}

	// 安装分布式服务发现插件
	if cb, ok := sb.composite.(InstallRegistry); ok {
		cb.InstallRegistry(servCtx)
	}
	if _, ok := servCtx.GetPluginBundle().Get(discovery.Name); !ok {
		etcd_discovery.Install(servCtx,
			etcd_discovery.Option{}.CustomAddresses(startupConf.GetString("etcd.address")),
			etcd_discovery.Option{}.CustomAuth(
				startupConf.GetString("etcd.username"),
				startupConf.GetString("etcd.password"),
			),
		)
	}

	// 安装分布式同步插件
	if cb, ok := sb.composite.(InstallDistSync); ok {
		cb.InstallDistSync(servCtx)
	}
	if _, ok := servCtx.GetPluginBundle().Get(dsync.Name); !ok {
		etcd_dsync.Install(servCtx,
			etcd_dsync.Option{}.CustomAddresses(startupConf.GetString("etcd.address")),
			etcd_dsync.Option{}.CustomAuth(
				startupConf.GetString("etcd.username"),
				startupConf.GetString("etcd.password"),
			),
		)
	}

	// 安装分布式服务插件
	if cb, ok := sb.composite.(InstallDistService); ok {
		cb.InstallDistService(servCtx)
	}
	if _, ok := servCtx.GetPluginBundle().Get(dserv.Name); !ok {
		dserv.Install(servCtx,
			dserv.Option{}.Version(startupConf.GetString(fmt.Sprintf("%s.version", sb.GetName()))),
			dserv.Option{}.Meta(startupConf.GetStringMapString(fmt.Sprintf("%s.meta", sb.GetName()))),
			dserv.Option{}.TTL(startupConf.GetDuration("service.ttl")),
			dserv.Option{}.FutureTimeout(startupConf.GetDuration("service.future_timeout")),
		)
	}

	// 安装RPC支持插件
	if cb, ok := sb.composite.(InstallRPC); ok {
		cb.InstallRPC(servCtx)
	}
	if _, ok := servCtx.GetPluginBundle().Get(rpc.Name); !ok {
		rpc.Install(servCtx)
	}

	// 安装分布式实体查询插件
	if cb, ok := sb.composite.(InstallDistEntityQuerier); ok {
		cb.InstallDistEntityQuerier(servCtx)
	}
	if _, ok := servCtx.GetPluginBundle().Get(dentq.Name); !ok {
		dentq.Install(servCtx,
			dentq.Option{}.CustomAddresses(startupConf.GetString("etcd.address")),
			dentq.Option{}.CustomAuth(
				startupConf.GetString("etcd.username"),
				startupConf.GetString("etcd.password"),
			),
		)
	}

	// 初始化回调
	if cb, ok := sb.composite.(LifecycleServiceInit); ok {
		cb.Init(servCtx)
	}

	return servCtx
}

func (sb *ServiceBehavior) GetName() string {
	return sb.name
}

func (sb *ServiceBehavior) GetStartupConf() *viper.Viper {
	return sb.app.startupConf
}
