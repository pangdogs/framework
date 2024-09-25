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
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/iface"
	"git.golaxy.org/core/utils/reinterpret"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/plugins/conf"
	"git.golaxy.org/framework/plugins/dentr"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/plugins/log/zap_log"
	"git.golaxy.org/framework/plugins/rpcstack"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type _RuntimeSettings struct {
	Name                 string
	PersistId            uid.Id
	AutoRecover          bool
	ReportError          chan error
	FPS                  float32
	ProcessQueueCapacity int
}

type iRuntimeGeneric interface {
	init(svcCtx service.Context, instance any)
	generate(settings _RuntimeSettings) core.Runtime
}

// RuntimeGeneric 运行时泛化类型
type RuntimeGeneric struct {
	svcInst  IServiceInstance
	instance any
}

func (r *RuntimeGeneric) init(svcCtx service.Context, instance any) {
	r.svcInst = reinterpret.Cast[IServiceInstance](svcCtx)
	r.instance = instance
}

func (r *RuntimeGeneric) generate(settings _RuntimeSettings) core.Runtime {
	wholeConf := conf.Using(r.svcInst).Whole()

	face := iface.Face[runtime.Context]{}

	if cb, ok := r.instance.(IRuntimeInstantiation); ok {
		face = iface.MakeFaceTReflectC[runtime.Context, IRuntimeInstance](cb.Instantiation())
	}

	iFrameLoopBeginCB, _ := face.Iface.(LifecycleRuntimeFrameLoopBegin)
	iFrameUpdateBeginCB, _ := face.Iface.(LifecycleRuntimeFrameUpdateBegin)
	iFrameUpdateEndCB, _ := face.Iface.(LifecycleRuntimeFrameUpdateEnd)
	iFrameLoopEndCB, _ := face.Iface.(LifecycleRuntimeFrameLoopEnd)
	iRunCallBeginCB, _ := face.Iface.(LifecycleRuntimeRunCallBegin)
	iRunCallEndCB, _ := face.Iface.(LifecycleRuntimeRunCallEnd)
	iRunGCBeginCB, _ := face.Iface.(LifecycleRuntimeRunGCBegin)
	iRunGCEndCB, _ := face.Iface.(LifecycleRuntimeRunGCEnd)

	frameLoopBeginCB, _ := r.instance.(LifecycleRuntimeFrameLoopBegin)
	frameUpdateBeginCB, _ := r.instance.(LifecycleRuntimeFrameUpdateBegin)
	frameUpdateEndCB, _ := r.instance.(LifecycleRuntimeFrameUpdateEnd)
	frameLoopEndCB, _ := r.instance.(LifecycleRuntimeFrameLoopEnd)
	runCallBeginCB, _ := r.instance.(LifecycleRuntimeRunCallBegin)
	runCallEndCB, _ := r.instance.(LifecycleRuntimeRunCallEnd)
	runGCBeginCB, _ := r.instance.(LifecycleRuntimeRunGCBegin)
	runGCEndCB, _ := r.instance.(LifecycleRuntimeRunGCEnd)

	rtCtx := runtime.NewContext(r.GetService(),
		runtime.With.Context.InstanceFace(face),
		runtime.With.Context.Name(settings.Name),
		runtime.With.Context.PersistId(settings.PersistId),
		runtime.With.Context.PanicHandling(settings.AutoRecover, settings.ReportError),
		runtime.With.Context.RunningHandler(generic.MakeDelegateAction2(func(rtCtx runtime.Context, state runtime.RunningState) {
			rtInst := reinterpret.Cast[IRuntimeInstance](rtCtx)

			switch state {
			case runtime.RunningState_Birth:
				if cb, ok := r.instance.(LifecycleRuntimeBirth); ok {
					cb.Birth(rtInst)
				}
				if cb, ok := rtInst.(LifecycleRuntimeBirth); ok {
					cb.Birth(rtInst)
				}
			case runtime.RunningState_Starting:
				if cb, ok := r.instance.(LifecycleRuntimeStarting); ok {
					cb.Starting(rtInst)
				}
				if cb, ok := rtInst.(LifecycleRuntimeStarting); ok {
					cb.Starting(rtInst)
				}
			case runtime.RunningState_Started:
				if cb, ok := r.instance.(LifecycleRuntimeStarted); ok {
					cb.Started(rtInst)
				}
				if cb, ok := rtInst.(LifecycleRuntimeStarted); ok {
					cb.Started(rtInst)
				}
			case runtime.RunningState_FrameLoopBegin:
				if cb := frameLoopBeginCB; cb != nil {
					cb.FrameLoopBegin(rtInst)
				}
				if cb := iFrameLoopBeginCB; cb != nil {
					cb.FrameLoopBegin(rtInst)
				}
			case runtime.RunningState_FrameUpdateBegin:
				if cb := frameUpdateBeginCB; cb != nil {
					cb.FrameUpdateBegin(rtInst)
				}
				if cb := iFrameUpdateBeginCB; cb != nil {
					cb.FrameUpdateBegin(rtInst)
				}
			case runtime.RunningState_FrameUpdateEnd:
				if cb := frameUpdateEndCB; cb != nil {
					cb.FrameUpdateEnd(rtInst)
				}
				if cb := iFrameUpdateEndCB; cb != nil {
					cb.FrameUpdateEnd(rtInst)
				}
			case runtime.RunningState_FrameLoopEnd:
				if cb := frameLoopEndCB; cb != nil {
					cb.FrameLoopEnd(rtInst)
				}
				if cb := iFrameLoopEndCB; cb != nil {
					cb.FrameLoopEnd(rtInst)
				}
			case runtime.RunningState_RunCallBegin:
				if cb := runCallBeginCB; cb != nil {
					cb.RunCallBegin(rtInst)
				}
				if cb := iRunCallBeginCB; cb != nil {
					cb.RunCallBegin(rtInst)
				}
			case runtime.RunningState_RunCallEnd:
				if cb := runCallEndCB; cb != nil {
					cb.RunCallEnd(rtInst)
				}
				if cb := iRunCallEndCB; cb != nil {
					cb.RunCallEnd(rtInst)
				}
			case runtime.RunningState_RunGCBegin:
				if cb := runGCBeginCB; cb != nil {
					cb.RunGCBegin(rtInst)
				}
				if cb := iRunGCBeginCB; cb != nil {
					cb.RunGCBegin(rtInst)
				}
			case runtime.RunningState_RunGCEnd:
				if cb := runGCEndCB; cb != nil {
					cb.RunGCEnd(rtInst)
				}
				if cb := iRunGCEndCB; cb != nil {
					cb.RunGCEnd(rtInst)
				}
			case runtime.RunningState_Terminating:
				if cb, ok := r.instance.(LifecycleRuntimeTerminating); ok {
					cb.Terminating(rtInst)
				}
				if cb, ok := rtInst.(LifecycleRuntimeTerminating); ok {
					cb.Terminating(rtInst)
				}
			case runtime.RunningState_Terminated:
				if cb, ok := r.instance.(LifecycleRuntimeTerminated); ok {
					cb.Terminated(rtInst)
				}
				if cb, ok := rtInst.(LifecycleRuntimeTerminated); ok {
					cb.Terminated(rtInst)
				}
			}
		})),
	)

	rtInst := reinterpret.Cast[IRuntimeInstance](rtCtx)

	installed := func(name string) bool {
		_, ok := rtInst.GetPluginBundle().Get(name)
		return ok
	}

	// 安装日志插件
	if !installed(log.Name) {
		if cb, ok := rtInst.(InstallRuntimeLogger); ok {
			cb.InstallLogger(rtInst)
		}
	}
	if !installed(log.Name) {
		if cb, ok := r.instance.(InstallRuntimeLogger); ok {
			cb.InstallLogger(rtInst)
		}
	}
	if !installed(log.Name) {
		if v, _ := r.svcInst.GetMemKV().Load("zap.logger"); v != nil {
			zap_log.Install(rtInst,
				zap_log.With.ZapLogger(v.(*zap.Logger)),
				zap_log.With.ServiceInfo(wholeConf.GetBool("log.service_info")),
				zap_log.With.RuntimeInfo(wholeConf.GetBool("log.runtime_info")),
			)
		}
	}

	// 安装RPC调用堆栈支持
	if !installed(rpcstack.Name) {
		if cb, ok := rtInst.(InstallRuntimeRPCStack); ok {
			cb.InstallRPCStack(rtInst)
		}
	}
	if !installed(rpcstack.Name) {
		if cb, ok := r.instance.(InstallRuntimeRPCStack); ok {
			cb.InstallRPCStack(rtInst)
		}
	}
	if !installed(rpcstack.Name) {
		rpcstack.Install(rtInst)
	}

	// 安装分布式实体支持插件
	if !installed(dentr.Name) {
		if cb, ok := rtInst.(InstallRuntimeDistEntityRegistry); ok {
			cb.InstallDistEntityRegistry(rtInst)
		}
	}
	if !installed(dentr.Name) {
		if cb, ok := r.instance.(InstallRuntimeDistEntityRegistry); ok {
			cb.InstallDistEntityRegistry(rtInst)
		}
	}
	if !installed(dentr.Name) {
		v, _ := r.GetService().GetMemKV().Load("etcd.client")
		cli, _ := v.(*etcdv3.Client)
		if cli == nil {
			panic("service memory kv etcd.client not existed")
		}

		dentr.Install(rtInst,
			dentr.With.EtcdClient(cli),
			dentr.With.TTL(wholeConf.GetDuration("service.dent_ttl")),
		)
	}

	// 组装完成回调回调
	if cb, ok := r.instance.(LifecycleRuntimeBuilt); ok {
		cb.Built(rtInst)
	}
	if cb, ok := rtInst.(LifecycleRuntimeBuilt); ok {
		cb.Built(rtInst)
	}

	return core.NewRuntime(rtInst,
		core.With.Runtime.Frame(func() runtime.Frame {
			if settings.FPS <= 0 {
				return nil
			}
			return runtime.NewFrame(
				runtime.With.Frame.TargetFPS(settings.FPS),
			)
		}()),
		core.With.Runtime.AutoRun(true),
		core.With.Runtime.ProcessQueueCapacity(settings.ProcessQueueCapacity),
	)
}

// GetService 获取服务
func (r *RuntimeGeneric) GetService() IServiceInstance {
	return r.svcInst
}
