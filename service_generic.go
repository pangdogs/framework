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
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/extension"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/iface"
	"git.golaxy.org/core/utils/reinterpret"
	"git.golaxy.org/framework/addins/broker"
	"git.golaxy.org/framework/addins/broker/nats_broker"
	"git.golaxy.org/framework/addins/conf"
	"git.golaxy.org/framework/addins/dentq"
	"git.golaxy.org/framework/addins/discovery"
	"git.golaxy.org/framework/addins/discovery/etcd_discovery"
	"git.golaxy.org/framework/addins/dsvc"
	"git.golaxy.org/framework/addins/dsync"
	"git.golaxy.org/framework/addins/dsync/etcd_dsync"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/addins/log/zap_log"
	"git.golaxy.org/framework/addins/rpc"
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type iServiceGeneric interface {
	init(startupConf *viper.Viper, name string, instance any)
	generate(ctx context.Context, no int) core.Service
}

// ServiceGeneric 服务泛化类型
type ServiceGeneric struct {
	once        sync.Once
	startupConf *viper.Viper
	name        string
	instance    any
}

func (s *ServiceGeneric) init(startupConf *viper.Viper, name string, instance any) {
	s.once.Do(func() {
		s.startupConf = startupConf
		s.name = name
		s.instance = instance
	})
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

	svcInstFace := iface.Face[service.Context]{}

	if cb, ok := s.instance.(IServiceInstantiation); ok {
		svcInstFace = iface.MakeFaceTReflectC[service.Context, IService](cb.Instantiate())
	} else {
		svcInstFace = iface.MakeFaceTReflectC[service.Context, IService](&Service{})
	}

	svcCtx := service.NewContext(
		service.With.InstanceFace(svcInstFace),
		service.With.Context(ctx),
		service.With.Name(s.GetName()),
		service.With.PanicHandling(autoRecover, reportError),
		service.With.RunningStatusChangedCB(func(svcCtx service.Context, status service.RunningStatus, args ...any) {
			svcInst := reinterpret.Cast[IService](svcCtx)

			switch status {
			case service.RunningStatus_Birth:
				if cb, ok := s.instance.(LifecycleServiceBirth); ok {
					cb.Birth(svcInst)
				}
				if cb, ok := svcInst.(LifecycleServiceBirth); ok {
					cb.Birth(svcInst)
				}
			case service.RunningStatus_Starting:
				if cb, ok := s.instance.(LifecycleServiceStarting); ok {
					cb.Starting(svcInst)
				}
				if cb, ok := svcInst.(LifecycleServiceStarting); ok {
					cb.Starting(svcInst)
				}
			case service.RunningStatus_Started:
				if cb, ok := s.instance.(LifecycleServiceStarted); ok {
					cb.Started(svcInst)
				}
				if cb, ok := svcInst.(LifecycleServiceStarted); ok {
					cb.Started(svcInst)
				}
			case service.RunningStatus_Terminating:
				if cb, ok := s.instance.(LifecycleServiceTerminating); ok {
					cb.Terminating(svcInst)
				}
				if cb, ok := svcInst.(LifecycleServiceTerminating); ok {
					cb.Terminating(svcInst)
				}
			case service.RunningStatus_Terminated:
				if cb, ok := s.instance.(LifecycleServiceTerminated); ok {
					cb.Terminated(svcInst)
				}
				if cb, ok := svcInst.(LifecycleServiceTerminated); ok {
					cb.Terminated(svcInst)
				}

				if v, ok := svcInst.GetMemKV().Load("zap.logger"); ok {
					v.(*zap.Logger).Sync()
				}

				if v, ok := svcInst.GetMemKV().Load("etcd.client"); ok {
					v.(*etcdv3.Client).Close()
				}
			case service.RunningStatus_ActivatingAddIn:
				addInStatus := args[0].(extension.AddInStatus)
				cacheCallPath(addInStatus.Name(), addInStatus.Reflected().Type())
				if cb, ok := s.instance.(LifecycleServiceAddInActivating); ok {
					cb.AddInActivating(svcInst, addInStatus)
				}
				if cb, ok := svcInst.(LifecycleServiceAddInActivating); ok {
					cb.AddInActivating(svcInst, addInStatus)
				}
			case service.RunningStatus_AddInActivated:
				addInStatus := args[0].(extension.AddInStatus)
				if cb, ok := s.instance.(LifecycleServiceAddInActivated); ok {
					cb.AddInActivated(svcInst, addInStatus)
				}
				if cb, ok := svcInst.(LifecycleServiceAddInActivated); ok {
					cb.AddInActivated(svcInst, addInStatus)
				}
			case service.RunningStatus_DeactivatingAddIn:
				addInStatus := args[0].(extension.AddInStatus)
				if cb, ok := s.instance.(LifecycleServiceAddInDeactivating); ok {
					cb.AddInDeactivating(svcInst, addInStatus)
				}
				if cb, ok := svcInst.(LifecycleServiceAddInDeactivating); ok {
					cb.AddInDeactivating(svcInst, addInStatus)
				}
			case service.RunningStatus_AddInDeactivated:
				addInStatus := args[0].(extension.AddInStatus)
				if cb, ok := s.instance.(LifecycleServiceAddInDeactivated); ok {
					cb.AddInDeactivated(svcInst, addInStatus)
				}
				if cb, ok := svcInst.(LifecycleServiceAddInDeactivated); ok {
					cb.AddInDeactivated(svcInst, addInStatus)
				}
			case service.RunningStatus_EntityPTDeclared:
				entityPT := args[0].(ec.EntityPT)
				cacheCallPath("", entityPT.InstanceRT())
				for i := range entityPT.CountComponents() {
					comp := entityPT.Component(i)
					cacheCallPath(comp.Name, comp.PT.InstanceRT())
				}
				if cb, ok := s.instance.(LifecycleServiceEntityPTDeclared); ok {
					cb.EntityPTDeclared(svcInst, entityPT)
				}
				if cb, ok := svcInst.(LifecycleServiceEntityPTDeclared); ok {
					cb.EntityPTDeclared(svcInst, entityPT)
				}
			case service.RunningStatus_EntityPTRedeclared:
				entityPT := args[0].(ec.EntityPT)
				cacheCallPath("", entityPT.InstanceRT())
				for i := range entityPT.CountComponents() {
					comp := entityPT.Component(i)
					cacheCallPath(comp.Name, comp.PT.InstanceRT())
				}
				if cb, ok := s.instance.(LifecycleServiceEntityPTRedeclared); ok {
					cb.EntityPTRedeclared(svcInst, entityPT)
				}
				if cb, ok := svcInst.(LifecycleServiceEntityPTRedeclared); ok {
					cb.EntityPTRedeclared(svcInst, entityPT)
				}
			case service.RunningStatus_EntityPTUndeclared:
				entityPT := args[0].(ec.EntityPT)
				if cb, ok := s.instance.(LifecycleServiceEntityPTUndeclared); ok {
					cb.EntityPTUndeclared(svcInst, entityPT)
				}
				if cb, ok := svcInst.(LifecycleServiceEntityPTUndeclared); ok {
					cb.EntityPTUndeclared(svcInst, entityPT)
				}
			}
		}),
	)

	svcInst := reinterpret.Cast[IService](svcCtx)
	cacheCallPath("", svcInst.GetReflected().Type())

	installed := func(name string) bool {
		_, ok := svcInst.GetAddInManager().Get(name)
		return ok
	}

	// 安装日志插件
	if !installed(log.Name) {
		if cb, ok := svcInst.(InstallServiceLogger); ok {
			cb.InstallLogger(svcInst)
		}
	}
	if !installed(log.Name) {
		if cb, ok := s.instance.(InstallServiceLogger); ok {
			cb.InstallLogger(svcInst)
		}
	}
	if !installed(log.Name) {
		level, err := zapcore.ParseLevel(startupConf.GetString("log.level"))
		if err != nil {
			exception.Panicf("%w: parse startup config [--log.level] = %q failed, %s", ErrFramework, startupConf.GetString("log.level"), err)
		}

		var filePath string

		if startupConf.GetString("log.dir") != "" {
			filePath = filepath.Join(startupConf.GetString("log.dir"), fmt.Sprintf("%s-%s-%d.log", strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0])), s.GetName(), no))
		}

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

		zap_log.Install(svcInst,
			zap_log.With.ZapLogger(zapLogger),
			zap_log.With.ServiceInfo(startupConf.GetBool("log.service_info")),
		)
	}

	// 安装配置插件
	if !installed(conf.Name) {
		if cb, ok := svcInst.(InstallServiceConfig); ok {
			cb.InstallConfig(svcInst)
		}
	}
	if !installed(conf.Name) {
		if cb, ok := s.instance.(InstallServiceConfig); ok {
			cb.InstallConfig(svcInst)
		}
	}
	if !installed(conf.Name) {
		conf.Install(svcInst,
			conf.With.Format(startupConf.GetString("conf.format")),
			conf.With.LocalPath(startupConf.GetString("conf.local_path")),
			conf.With.Remote(
				startupConf.GetString("conf.remote_provider"),
				startupConf.GetString("conf.remote_endpoint"),
				startupConf.GetString("conf.remote_path"),
			),
			conf.With.AutoHotFix(startupConf.GetBool("conf.auto_update")),
			conf.With.MergeConf(startupConf),
		)
	}

	// 安装消息队列中间件插件
	if !installed(broker.Name) {
		if cb, ok := svcInst.(InstallServiceBroker); ok {
			cb.InstallBroker(svcInst)
		}
	}
	if !installed(broker.Name) {
		if cb, ok := s.instance.(InstallServiceBroker); ok {
			cb.InstallBroker(svcInst)
		}
	}
	if !installed(broker.Name) {
		nats_broker.Install(svcInst,
			nats_broker.With.CustomAddresses(startupConf.GetString("nats.address")),
			nats_broker.With.CustomAuth(
				startupConf.GetString("nats.username"),
				startupConf.GetString("nats.password"),
			),
		)
	}

	// 安装服务发现插件
	if !installed(discovery.Name) {
		if cb, ok := svcInst.(InstallServiceRegistry); ok {
			cb.InstallRegistry(svcInst)
		}
	}
	if !installed(discovery.Name) {
		if cb, ok := s.instance.(InstallServiceRegistry); ok {
			cb.InstallRegistry(svcInst)
		}
	}
	if !installed(discovery.Name) {
		etcd_discovery.Install(svcInst,
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
		if cb, ok := svcInst.(InstallServiceDistSync); ok {
			cb.InstallDistSync(svcInst)
		}
	}
	if !installed(dsync.Name) {
		if cb, ok := s.instance.(InstallServiceDistSync); ok {
			cb.InstallDistSync(svcInst)
		}
	}
	if !installed(dsync.Name) {
		etcd_dsync.Install(svcInst,
			etcd_dsync.With.CustomAddresses(startupConf.GetString("etcd.address")),
			etcd_dsync.With.CustomAuth(
				startupConf.GetString("etcd.username"),
				startupConf.GetString("etcd.password"),
			),
		)
	}

	// 安装分布式服务插件
	if !installed(dsvc.Name) {
		if cb, ok := svcInst.(InstallServiceDistService); ok {
			cb.InstallDistService(svcInst)
		}
	}
	if !installed(dsvc.Name) {
		if cb, ok := s.instance.(InstallServiceDistService); ok {
			cb.InstallDistService(svcInst)
		}
	}
	if !installed(dsvc.Name) {
		dsvc.Install(svcInst,
			dsvc.With.Version(startupConf.GetString("service.version")),
			dsvc.With.Meta(startupConf.GetStringMapString("service.meta")),
			dsvc.With.FutureTimeout(startupConf.GetDuration("service.future_timeout")),
		)
	}

	// 安装分布式实体查询插件
	if !installed(dentq.Name) {
		if cb, ok := svcInst.(InstallServiceDistEntityQuerier); ok {
			cb.InstallDistEntityQuerier(svcInst)
		}
	}
	if !installed(dentq.Name) {
		if cb, ok := s.instance.(InstallServiceDistEntityQuerier); ok {
			cb.InstallDistEntityQuerier(svcInst)
		}
	}
	if !installed(dentq.Name) {
		dentq.Install(svcInst,
			dentq.With.CustomAddresses(startupConf.GetString("etcd.address")),
			dentq.With.CustomAuth(
				startupConf.GetString("etcd.username"),
				startupConf.GetString("etcd.password"),
			),
		)
	}

	// 安装RPC支持插件
	if !installed(rpc.Name) {
		if cb, ok := svcInst.(InstallServiceRPC); ok {
			cb.InstallRPC(svcInst)
		}
	}
	if !installed(rpc.Name) {
		if cb, ok := s.instance.(InstallServiceRPC); ok {
			cb.InstallRPC(svcInst)
		}
	}
	if !installed(rpc.Name) {
		rpc.Install(svcInst)
	}

	// 组装完成回调
	if cb, ok := s.instance.(LifecycleServiceBuilt); ok {
		cb.Built(svcInst)
	}
	if cb, ok := svcInst.(LifecycleServiceBuilt); ok {
		cb.Built(svcInst)
	}

	// 延迟连接etcd
	memKV.Store("etcd.lazy_conn", sync.OnceValue(func() *etcdv3.Client {
		cli, err := etcdv3.New(etcdv3.Config{
			Endpoints: []string{startupConf.GetString("etcd.address")},
			Username:  startupConf.GetString("etcd.username"),
			Password:  startupConf.GetString("etcd.password"),
		})
		if err != nil {
			exception.Panicf("%w: new etcd client failed, %s", ErrFramework, err)
		}
		memKV.Store("etcd.client", cli)
		return cli
	}))

	// 自动恢复时，打印panic信息
	if svcInst.GetAutoRecover() && svcInst.GetReportError() != nil {
		go func() {
			for {
				select {
				case err := <-svcInst.GetReportError():
					log.Errorf(svcInst, "recover:\n%s", err)
				case <-svcInst.Done():
					return
				}
			}
		}()
	}

	// 创建服务
	return core.NewService(svcInst)
}

// GetName 获取服务名称
func (s *ServiceGeneric) GetName() string {
	return s.name
}

// GetStartupConf 获取启动参数配置
func (s *ServiceGeneric) GetStartupConf() *viper.Viper {
	return s.startupConf
}
