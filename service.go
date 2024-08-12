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

package framework

import (
	"context"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/pt"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/iface"
	"git.golaxy.org/core/utils/reinterpret"
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
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
)

type iServiceGeneric interface {
	init(startupConf *viper.Viper, name string, composite any)
	generate(ctx context.Context, no int) core.Service
}

// ServiceGenericT 服务泛化类型实例化
type ServiceGenericT[T any] struct {
	ServiceGeneric
}

// Instantiation 实例化
func (s *ServiceGenericT[T]) Instantiation() IServiceInstance {
	return reflect.New(reflect.TypeFor[T]()).Interface().(IServiceInstance)
}

// ServiceGeneric 服务泛化类型
type ServiceGeneric struct {
	startupConf *viper.Viper
	name        string
	composite   any
}

func (s *ServiceGeneric) init(startupConf *viper.Viper, name string, composite any) {
	s.startupConf = startupConf
	s.name = name
	s.composite = composite
}

func (s *ServiceGeneric) generate(ctx context.Context, no int) core.Service {
	startupConf := s.GetStartupConf()

	memKV := &sync.Map{}
	memKV.Store("startup.no", no)
	memKV.Store("startup.conf", startupConf)

	ctx = context.WithValue(ctx, "mem_kv", memKV)

	autoRecover := startupConf.GetBool("service.auto_recover")
	var reportError chan error

	if autoRecover {
		reportError = make(chan error, 128)
	}

	face := iface.Face[service.Context]{}

	if cb, ok := s.composite.(IServiceInstantiation); ok {
		face = iface.MakeFaceTReflectC[service.Context, IServiceInstance](cb.Instantiation())
	}

	servCtx := service.NewContext(
		service.With.CompositeFace(face),
		service.With.Context(ctx),
		service.With.Name(s.GetName()),
		service.With.PanicHandling(autoRecover, reportError),
		service.With.EntityLib(pt.NewEntityLib(pt.DefaultComponentLib())),
		service.With.RunningHandler(generic.MakeDelegateAction2(func(ctx service.Context, state service.RunningState) {
			inst := reinterpret.Cast[IServiceInstance](ctx)

			switch state {
			case service.RunningState_Birth:
				if cb, ok := s.composite.(LifecycleServiceBirth); ok {
					cb.Birth(inst)
				}
				if cb, ok := inst.(LifecycleServiceBirth); ok {
					cb.Birth(inst)
				}
			case service.RunningState_Starting:
				if cb, ok := s.composite.(LifecycleServiceStarting); ok {
					cb.Starting(inst)
				}
				if cb, ok := inst.(LifecycleServiceStarting); ok {
					cb.Starting(inst)
				}
			case service.RunningState_Started:
				if cb, ok := s.composite.(LifecycleServiceStarted); ok {
					cb.Started(inst)
				}
				if cb, ok := inst.(LifecycleServiceStarted); ok {
					cb.Started(inst)
				}
			case service.RunningState_Terminating:
				if cb, ok := s.composite.(LifecycleServiceTerminating); ok {
					cb.Terminating(inst)
				}
				if cb, ok := inst.(LifecycleServiceTerminating); ok {
					cb.Terminating(inst)
				}
			case service.RunningState_Terminated:
				if cb, ok := s.composite.(LifecycleServiceTerminated); ok {
					cb.Terminated(inst)
				}
				if cb, ok := inst.(LifecycleServiceTerminated); ok {
					cb.Terminated(inst)
				}

				if v, ok := inst.GetMemKV().Load("zap.logger"); ok {
					v.(*zap.Logger).Sync()
				}

				if v, ok := inst.GetMemKV().Load("etcd.client"); ok {
					v.(*etcdv3.Client).Close()
				}
			}
		})),
	)

	servInst := reinterpret.Cast[IServiceInstance](servCtx)

	installed := func(name string) bool {
		_, ok := servInst.GetPluginBundle().Get(name)
		return ok
	}

	// 安装日志插件
	if !installed(log.Name) {
		if cb, ok := servInst.(InstallServiceLogger); ok {
			cb.InstallLogger(servInst)
		}
	}
	if !installed(log.Name) {
		if cb, ok := s.composite.(InstallServiceLogger); ok {
			cb.InstallLogger(servInst)
		}
	}
	if !installed(log.Name) {
		level, err := zapcore.ParseLevel(startupConf.GetString("log.level"))
		if err != nil {
			panic(fmt.Errorf("parse startup config [--log.level] = %q failed, %s", startupConf.GetString("log.level"), err))
		}

		filePath := filepath.Join(startupConf.GetString("log.dir"), fmt.Sprintf("%s-%s-%d.log", strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0])), s.GetName(), no))

		var zapLogger *zap.Logger
		var zapAtomicLevel zap.AtomicLevel

		switch startupConf.GetString("log.format") {
		case "json":
			zapLogger, zapAtomicLevel = zap_log.NewJsonZapLogger(
				level,
				filePath,
				startupConf.GetInt("log.size"),
				startupConf.GetBool("log.stdout"),
				startupConf.GetBool("log.development"),
			)
		default:
			zapLogger, zapAtomicLevel = zap_log.NewConsoleZapLogger(
				level,
				"\t",
				filePath,
				startupConf.GetInt("log.size"),
				startupConf.GetBool("log.stdout"),
				startupConf.GetBool("log.development"),
			)
		}

		memKV.Store("zap.logger", zapLogger)
		memKV.Store("zap.atomic_level", zapAtomicLevel)

		zap_log.Install(servInst,
			zap_log.With.ZapLogger(zapLogger),
			zap_log.With.ServiceInfo(startupConf.GetBool("log.service_info")),
		)
	}

	// 安装配置插件
	if !installed(conf.Name) {
		if cb, ok := servInst.(InstallServiceConfig); ok {
			cb.InstallConfig(servInst)
		}
	}
	if !installed(conf.Name) {
		if cb, ok := s.composite.(InstallServiceConfig); ok {
			cb.InstallConfig(servInst)
		}
	}
	if !installed(conf.Name) {
		conf.Install(servInst,
			conf.With.Format(startupConf.GetString("conf.format")),
			conf.With.LocalPath(startupConf.GetString("conf.local_path")),
			conf.With.Remote(
				startupConf.GetString("conf.remote_provider"),
				startupConf.GetString("conf.remote_endpoint"),
				startupConf.GetString("conf.remote_path"),
			),
			conf.With.AutoUpdate(startupConf.GetBool("conf.auto_update")),
			conf.With.MergeConf(startupConf),
		)
	}

	// 安装消息队列中间件插件
	if !installed(broker.Name) {
		if cb, ok := servInst.(InstallServiceBroker); ok {
			cb.InstallBroker(servInst)
		}
	}
	if !installed(broker.Name) {
		if cb, ok := s.composite.(InstallServiceBroker); ok {
			cb.InstallBroker(servInst)
		}
	}
	if !installed(broker.Name) {
		nats_broker.Install(servInst,
			nats_broker.With.CustomAddresses(startupConf.GetString("nats.address")),
			nats_broker.With.CustomAuth(
				startupConf.GetString("nats.username"),
				startupConf.GetString("nats.password"),
			),
		)
	}

	// 安装服务发现插件
	if !installed(discovery.Name) {
		if cb, ok := servInst.(InstallServiceRegistry); ok {
			cb.InstallRegistry(servInst)
		}
	}
	if !installed(discovery.Name) {
		if cb, ok := s.composite.(InstallServiceRegistry); ok {
			cb.InstallRegistry(servInst)
		}
	}
	if !installed(discovery.Name) {
		etcd_discovery.Install(servInst,
			etcd_discovery.With.TTL(startupConf.GetDuration("service.ttl"), true),
			etcd_discovery.With.CustomAddresses(startupConf.GetString("etcd.address")),
			etcd_discovery.With.CustomAuth(
				startupConf.GetString("etcd.username"),
				startupConf.GetString("etcd.password"),
			),
		)
	}

	// 安装分布式同步插件
	if !installed(dsync.Name) {
		if cb, ok := servInst.(InstallServiceDistSync); ok {
			cb.InstallDistSync(servInst)
		}
	}
	if !installed(dsync.Name) {
		if cb, ok := s.composite.(InstallServiceDistSync); ok {
			cb.InstallDistSync(servInst)
		}
	}
	if !installed(dsync.Name) {
		etcd_dsync.Install(servInst,
			etcd_dsync.With.CustomAddresses(startupConf.GetString("etcd.address")),
			etcd_dsync.With.CustomAuth(
				startupConf.GetString("etcd.username"),
				startupConf.GetString("etcd.password"),
			),
		)
	}

	// 安装分布式服务插件
	if !installed(dserv.Name) {
		if cb, ok := servInst.(InstallServiceDistService); ok {
			cb.InstallDistService(servInst)
		}
	}
	if !installed(dserv.Name) {
		if cb, ok := s.composite.(InstallServiceDistService); ok {
			cb.InstallDistService(servInst)
		}
	}
	if !installed(dserv.Name) {
		dserv.Install(servInst,
			dserv.With.Version(startupConf.GetString("service.version")),
			dserv.With.Meta(startupConf.GetStringMapString("service.meta")),
			dserv.With.FutureTimeout(startupConf.GetDuration("service.future_timeout")),
		)
	}

	// 安装分布式实体查询插件
	if !installed(dentq.Name) {
		if cb, ok := servInst.(InstallServiceDistEntityQuerier); ok {
			cb.InstallDistEntityQuerier(servInst)
		}
	}
	if !installed(dentq.Name) {
		if cb, ok := s.composite.(InstallServiceDistEntityQuerier); ok {
			cb.InstallDistEntityQuerier(servInst)
		}
	}
	if !installed(dentq.Name) {
		dentq.Install(servInst,
			dentq.With.CustomAddresses(startupConf.GetString("etcd.address")),
			dentq.With.CustomAuth(
				startupConf.GetString("etcd.username"),
				startupConf.GetString("etcd.password"),
			),
		)
	}

	// 安装RPC支持插件
	if !installed(rpc.Name) {
		if cb, ok := servInst.(InstallServiceRPC); ok {
			cb.InstallRPC(servInst)
		}
	}
	if !installed(rpc.Name) {
		if cb, ok := s.composite.(InstallServiceRPC); ok {
			cb.InstallRPC(servInst)
		}
	}
	if !installed(rpc.Name) {
		rpc.Install(servInst)
	}

	// etcd连接初始化函数
	memKV.Store("etcd.init_client", sync.OnceFunc(func() {
		cli, err := etcdv3.New(etcdv3.Config{
			Endpoints: []string{startupConf.GetString("etcd.address")},
			Username:  startupConf.GetString("etcd.username"),
			Password:  startupConf.GetString("etcd.password"),
		})
		if err != nil {
			panic(fmt.Errorf("new etcd client failed, %s", err))
		}
		memKV.Store("etcd.client", cli)
	}))

	// 组装完成回调回调
	if cb, ok := s.composite.(LifecycleServiceBuilt); ok {
		cb.Built(servInst)
	}
	if cb, ok := servInst.(LifecycleServiceBuilt); ok {
		cb.Built(servInst)
	}

	// 自动恢复时，打印panic信息
	if servInst.GetAutoRecover() && servInst.GetReportError() != nil {
		go func() {
			for {
				select {
				case err := <-servInst.GetReportError():
					log.Errorf(servInst, "recover:\n%s", err)
				case <-servInst.Done():
					return
				}
			}
		}()
	}

	return core.NewService(servInst)
}

// GetName 获取服务名称
func (s *ServiceGeneric) GetName() string {
	return s.name
}

// GetStartupConf 获取启动参数配置
func (s *ServiceGeneric) GetStartupConf() *viper.Viper {
	return s.startupConf
}
