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
	"os"
	"sync"

	"git.golaxy.org/core"
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/extension"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/iface"
	"git.golaxy.org/core/utils/reinterpret"
	. "git.golaxy.org/framework/addins"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type iServiceAssembler interface {
	init(conf *viper.Viper, cmd *cobra.Command, name string, instance any)
	assemble(ctx context.Context, replicaNo int) core.Service
}

const (
	memConf              = "conf"
	memCmd               = "cmd"
	memReplicaNo         = "replica-no"
	memLogger            = "logger"
	memLoggerAtomicLevel = "logger-atomic-level"
	memEtcdClientOnce    = "etcd-client-once"
	memEtcdClient        = "etcd-client"
)

// ServiceAssembler 服务装配器
type ServiceAssembler struct {
	conf     *viper.Viper
	cmd      *cobra.Command
	name     string
	instance any
}

func (s *ServiceAssembler) init(conf *viper.Viper, cmd *cobra.Command, name string, instance any) {
	s.conf = conf
	s.cmd = cmd
	s.name = name
	s.instance = instance
}

func (s *ServiceAssembler) assemble(ctx context.Context, replicaNo int) core.Service {
	svcInstFace := iface.Face[service.Context]{}

	if cb, ok := s.instance.(IServiceInstantiator); ok {
		svcInstFace = iface.NewFaceTReflectC[service.Context, IService](cb.Instantiate())
	} else {
		svcInstFace = iface.NewFaceTReflectC[service.Context, IService](&ServiceBehavior{})
	}

	autoRecover := s.conf.GetBool("service.auto_recover")
	var reportError chan error

	if autoRecover {
		reportError = make(chan error, 128)
	}

	svcCtx := service.NewContext(
		service.With.InstanceFace(svcInstFace),
		service.With.Context(ctx),
		service.With.Name(s.name),
		service.With.PanicHandling(autoRecover, reportError),
		service.With.RunningEventCB(func(svcCtx service.Context, runningEvent service.RunningEvent, args ...any) {
			svcInst := reinterpret.Cast[IService](svcCtx)

			switch runningEvent {
			case service.RunningEvent_Birth:
				svcInst.(iService).getRuntimeAssembler().init(svcInst, svcInst.(iService).getRuntimeAssembler())
				cacheCallPath("", svcInst.Reflected().Type())

				svcInst.Memory().Store(memReplicaNo, replicaNo)

				s.initConf(svcInst)
				s.initLogger(svcInst)

				if cb, ok := s.instance.(LifecycleServiceBirth); ok {
					cb.OnBirth(svcInst)
				}
				if cb, ok := svcInst.(LifecycleServiceBirth); ok {
					cb.OnBirth(svcInst)
				}

				s.installAddIns(svcInst)

				if cb, ok := s.instance.(LifecycleServiceBuilt); ok {
					cb.OnBuilt(svcInst)
				}
				if cb, ok := svcInst.(LifecycleServiceBuilt); ok {
					cb.OnBuilt(svcInst)
				}

				svcInst.Memory().Store(memEtcdClientOnce, sync.OnceValue(func() *etcdv3.Client {
					cli, err := etcdv3.New(etcdv3.Config{
						Endpoints: []string{s.conf.GetString("etcd.address")},
						Username:  s.conf.GetString("etcd.username"),
						Password:  s.conf.GetString("etcd.password"),
					})
					if err != nil {
						exception.Panicf("%w: new etcd client failed, %s", ErrFramework, err)
					}
					svcInst.Memory().Store(memEtcdClient, cli)
					return cli
				}))

				if svcInst.AutoRecover() && svcInst.ReportError() != nil {
					go func() {
						for {
							select {
							case err := <-svcInst.ReportError():
								svcInst.L().Error("[Recovery from panic]", zap.Error(err))
							case <-svcInst.Done():
								return
							}
						}
					}()
				}
			case service.RunningEvent_Starting:
				if cb, ok := s.instance.(LifecycleServiceStarting); ok {
					cb.OnStarting(svcInst)
				}
				if cb, ok := svcInst.(LifecycleServiceStarting); ok {
					cb.OnStarting(svcInst)
				}
			case service.RunningEvent_Started:
				if !svcInst.(iService).getStarted().CompareAndSwap(false, true) {
					exception.Panicf("%w: already started", ErrFramework)
				}

				// 服务上线
				svcInst.DistService().BringUp()

				if cb, ok := s.instance.(LifecycleServiceStarted); ok {
					cb.OnStarted(svcInst)
				}
				if cb, ok := svcInst.(LifecycleServiceStarted); ok {
					cb.OnStarted(svcInst)
				}
			case service.RunningEvent_Terminating:
				if cb, ok := s.instance.(LifecycleServiceTerminating); ok {
					cb.OnTerminating(svcInst)
				}
				if cb, ok := svcInst.(LifecycleServiceTerminating); ok {
					cb.OnTerminating(svcInst)
				}
			case service.RunningEvent_Terminated:
				if cb, ok := s.instance.(LifecycleServiceTerminated); ok {
					cb.OnTerminated(svcInst)
				}
				if cb, ok := svcInst.(LifecycleServiceTerminated); ok {
					cb.OnTerminated(svcInst)
				}
				{
					v, _ := svcInst.Memory().Load(memLogger)
					if logger, ok := v.(*zap.Logger); ok {
						logger.Sync()
					}
				}
				{
					v, _ := svcInst.Memory().Load(memEtcdClient)
					if cli, ok := v.(*etcdv3.Client); ok {
						cli.Close()
					}
				}
			case service.RunningEvent_AddInActivating:
				addInStatus := args[0].(extension.AddInStatus)
				cacheCallPath(addInStatus.Name(), addInStatus.Reflected().Type())
				if cb, ok := s.instance.(LifecycleServiceAddInActivating); ok {
					cb.OnAddInActivating(svcInst, addInStatus)
				}
				if cb, ok := svcInst.(LifecycleServiceAddInActivating); ok {
					cb.OnAddInActivating(svcInst, addInStatus)
				}
			case service.RunningEvent_AddInActivated:
				addInStatus := args[0].(extension.AddInStatus)
				if cb, ok := s.instance.(LifecycleServiceAddInActivated); ok {
					cb.OnAddInActivated(svcInst, addInStatus)
				}
				if cb, ok := svcInst.(LifecycleServiceAddInActivated); ok {
					cb.OnAddInActivated(svcInst, addInStatus)
				}
			case service.RunningEvent_AddInDeactivating:
				addInStatus := args[0].(extension.AddInStatus)
				if cb, ok := s.instance.(LifecycleServiceAddInDeactivating); ok {
					cb.OnAddInDeactivating(svcInst, addInStatus)
				}
				if cb, ok := svcInst.(LifecycleServiceAddInDeactivating); ok {
					cb.OnAddInDeactivating(svcInst, addInStatus)
				}
			case service.RunningEvent_AddInDeactivated:
				addInStatus := args[0].(extension.AddInStatus)
				if cb, ok := s.instance.(LifecycleServiceAddInDeactivated); ok {
					cb.OnAddInDeactivated(svcInst, addInStatus)
				}
				if cb, ok := svcInst.(LifecycleServiceAddInDeactivated); ok {
					cb.OnAddInDeactivated(svcInst, addInStatus)
				}
			case service.RunningEvent_EntityPTDeclared:
				entityPT := args[0].(ec.EntityPT)
				cacheCallPath("", entityPT.InstanceRT())
				for i := range entityPT.CountComponents() {
					comp := entityPT.GetComponent(i)
					cacheCallPath(comp.Name, comp.PT.InstanceRT())
				}
				if cb, ok := s.instance.(LifecycleServiceEntityPTDeclared); ok {
					cb.OnEntityPTDeclared(svcInst, entityPT)
				}
				if cb, ok := svcInst.(LifecycleServiceEntityPTDeclared); ok {
					cb.OnEntityPTDeclared(svcInst, entityPT)
				}
			case service.RunningEvent_ComponentPTDeclared:
				compPT := args[0].(ec.ComponentPT)
				if cb, ok := s.instance.(LifecycleServiceComponentPTDeclared); ok {
					cb.OnComponentPTDeclared(svcInst, compPT)
				}
				if cb, ok := svcInst.(LifecycleServiceComponentPTDeclared); ok {
					cb.OnComponentPTDeclared(svcInst, compPT)
				}
			case service.RunningEvent_EntityRegistered:
				entity := args[0].(ec.ConcurrentEntity)
				if cb, ok := s.instance.(LifecycleServiceEntityRegistered); ok {
					cb.OnEntityRegistered(svcInst, entity)
				}
				if cb, ok := svcInst.(LifecycleServiceEntityRegistered); ok {
					cb.OnEntityRegistered(svcInst, entity)
				}
			case service.RunningEvent_EntityDeregistered:
				entity := args[0].(ec.ConcurrentEntity)
				if cb, ok := s.instance.(LifecycleServiceEntityDeregistered); ok {
					cb.OnEntityDeregistered(svcInst, entity)
				}
				if cb, ok := svcInst.(LifecycleServiceEntityDeregistered); ok {
					cb.OnEntityDeregistered(svcInst, entity)
				}
			}
		}),
	)

	return core.NewService(svcCtx)
}

func (s *ServiceAssembler) initConf(svcInst IService) {
	conf := viper.New()
	conf.MergeConfigMap(s.conf.AllSettings())

	svcInst.Memory().Store(memConf, conf)
	svcInst.Memory().Store(memCmd, s.cmd)
}

func (s *ServiceAssembler) initLogger(svcInst IService) {
	conf := svcInst.AppConf()

	level, err := zapcore.ParseLevel(conf.GetString("log.level"))
	if err != nil {
		exception.Panicf("%w: parse log.level:%q failed, %s", ErrFramework, conf.GetString("log.level"), err)
	}

	var encoderConf zapcore.EncoderConfig
	switch conf.GetString("log.encoder") {
	case "production":
		encoderConf = zap.NewProductionEncoderConfig()
	case "development":
		encoderConf = zap.NewDevelopmentEncoderConfig()
	default:
		exception.Panicf("%w: unknown log.encoder:%q", ErrFramework, conf.GetString("log.encoder"))
	}

	var encoder zapcore.Encoder
	switch conf.GetString("log.format") {
	case "console":
		encoder = zapcore.NewConsoleEncoder(encoderConf)
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConf)
	default:
		exception.Panicf("%w: unknown log.format:%q", ErrFramework, conf.GetString("log.format"))
	}

	var logger *zap.Logger
	atomicLevel := zap.NewAtomicLevelAt(level)

	if conf.GetBool("log.async") {
		logger = zap.New(
			zapcore.NewCore(
				encoder,
				&zapcore.BufferedWriteSyncer{
					WS:            zapcore.AddSync(os.Stdout),
					Size:          conf.GetInt("log.buffer_size"),
					FlushInterval: conf.GetDuration("log.flush_interval"),
				},
				atomicLevel,
			),
			zap.AddCaller(),
			zap.AddStacktrace(zap.DPanicLevel),
		)
	} else {
		logger = zap.New(
			zapcore.NewCore(
				encoder,
				zapcore.AddSync(os.Stdout),
				atomicLevel,
			),
			zap.AddCaller(),
			zap.AddStacktrace(zap.DPanicLevel),
		)
	}

	svcInst.Memory().Store(memLogger, logger)
	svcInst.Memory().Store(memLoggerAtomicLevel, atomicLevel)
}

func (s *ServiceAssembler) installAddIns(svcInst IService) {
	conf := svcInst.AppConf()

	installed := func(name string) bool {
		_, ok := svcInst.AddInManager().GetStatusByName(name)
		return ok
	}
	requireInstalled := func(name string) {
		if !installed(name) {
			exception.Panicf("%w: service add-in %q not installed", ErrFramework, name)
		}
	}

	// 安装日志插件
	if !installed(Log.Name) {
		if cb, ok := svcInst.(InstallServiceLogger); ok {
			cb.InstallLogger(svcInst)
		}
	}
	if !installed(Log.Name) {
		if cb, ok := s.instance.(InstallServiceLogger); ok {
			cb.InstallLogger(svcInst)
		}
	}
	if !installed(Log.Name) {
		Log.Install(svcInst,
			LogWith.Logger(svcInst.L()),
		)
	}
	requireInstalled(Log.Name)

	// 安装配置插件
	if !installed(Conf.Name) {
		if cb, ok := svcInst.(InstallServiceConfig); ok {
			cb.InstallConfig(svcInst)
		}
	}
	if !installed(Conf.Name) {
		if cb, ok := s.instance.(InstallServiceConfig); ok {
			cb.InstallConfig(svcInst)
		}
	}
	if !installed(Conf.Name) {
		Conf.Install(svcInst,
			ConfWith.Vipper(conf),
		)
	}
	requireInstalled(Conf.Name)

	// 安装消息队列中间件插件
	if !installed(Broker.Name) {
		if cb, ok := svcInst.(InstallServiceBroker); ok {
			cb.InstallBroker(svcInst)
		}
	}
	if !installed(Broker.Name) {
		if cb, ok := s.instance.(InstallServiceBroker); ok {
			cb.InstallBroker(svcInst)
		}
	}
	if !installed(Broker.Name) {
		BrokerNats.Install(svcInst,
			BrokerNatsWith.CustomAddresses(conf.GetString("nats.address")),
			BrokerNatsWith.CustomAuth(
				conf.GetString("nats.username"),
				conf.GetString("nats.password"),
			),
		)
	}
	requireInstalled(Broker.Name)

	// 安装服务发现插件
	if !installed(Discovery.Name) {
		if cb, ok := svcInst.(InstallServiceRegistry); ok {
			cb.InstallRegistry(svcInst)
		}
	}
	if !installed(Discovery.Name) {
		if cb, ok := s.instance.(InstallServiceRegistry); ok {
			cb.InstallRegistry(svcInst)
		}
	}
	if !installed(Discovery.Name) {
		DiscoveryEtcd.Install(svcInst,
			DiscoveryEtcdWith.CustomAddresses(conf.GetString("etcd.address")),
			DiscoveryEtcdWith.CustomAuth(
				conf.GetString("etcd.username"),
				conf.GetString("etcd.password"),
			),
		)
	}
	requireInstalled(Discovery.Name)

	// 安装分布式同步插件
	if !installed(Dsync.Name) {
		if cb, ok := svcInst.(InstallServiceDistSync); ok {
			cb.InstallDistSync(svcInst)
		}
	}
	if !installed(Dsync.Name) {
		if cb, ok := s.instance.(InstallServiceDistSync); ok {
			cb.InstallDistSync(svcInst)
		}
	}
	if !installed(Dsync.Name) {
		DsyncEtcd.Install(svcInst,
			DsyncEtcdWith.CustomAddresses(conf.GetString("etcd.address")),
			DsyncEtcdWith.CustomAuth(
				conf.GetString("etcd.username"),
				conf.GetString("etcd.password"),
			),
		)
	}
	requireInstalled(Dsync.Name)

	// 安装分布式服务插件
	if !installed(Dsvc.Name) {
		if cb, ok := svcInst.(InstallServiceDistService); ok {
			cb.InstallDistService(svcInst)
		}
	}
	if !installed(Dsvc.Name) {
		if cb, ok := s.instance.(InstallServiceDistService); ok {
			cb.InstallDistService(svcInst)
		}
	}
	if !installed(Dsvc.Name) {
		Dsvc.Install(svcInst,
			DsvcWith.Version(conf.GetString("service.version")),
			DsvcWith.Meta(conf.GetStringMapString("service.meta")),
			DsvcWith.RegistrationTTL(conf.GetDuration("service.ttl")),
			DsvcWith.FutureTimeout(conf.GetDuration("service.future_timeout")),
		)
	}
	requireInstalled(Dsvc.Name)

	// 安装分布式实体查询插件
	if !installed(Dentq.Name) {
		if cb, ok := svcInst.(InstallServiceDistEntityQuerier); ok {
			cb.InstallDistEntityQuerier(svcInst)
		}
	}
	if !installed(Dentq.Name) {
		if cb, ok := s.instance.(InstallServiceDistEntityQuerier); ok {
			cb.InstallDistEntityQuerier(svcInst)
		}
	}
	if !installed(Dentq.Name) {
		Dentq.Install(svcInst,
			DentqWith.CustomAddresses(conf.GetString("etcd.address")),
			DentqWith.CustomAuth(
				conf.GetString("etcd.username"),
				conf.GetString("etcd.password"),
			),
		)
	}
	requireInstalled(Dentq.Name)

	// 安装RPC支持插件
	if !installed(RPC.Name) {
		if cb, ok := svcInst.(InstallServiceRPC); ok {
			cb.InstallRPC(svcInst)
		}
	}
	if !installed(RPC.Name) {
		if cb, ok := s.instance.(InstallServiceRPC); ok {
			cb.InstallRPC(svcInst)
		}
	}
	if !installed(RPC.Name) {
		RPC.Install(svcInst)
	}
	requireInstalled(RPC.Name)
}
