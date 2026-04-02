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
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/extension"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/assertion"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/iface"
	"git.golaxy.org/core/utils/reinterpret"
	"git.golaxy.org/core/utils/uid"
	. "git.golaxy.org/framework/addins"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type _RuntimeSettings struct {
	name                            string
	persistId                       uid.Id
	autoRecover                     bool
	reportError                     chan error
	continueOnActivatingEntityPanic bool
	enableFrame                     bool
	fps                             float64
	autoInjection                   bool
}

type iRuntimeAssembler interface {
	init(svcCtx service.Context, instance any)
	assemble(settings _RuntimeSettings) core.Runtime
}

// RuntimeAssembler 运行时装配器
type RuntimeAssembler struct {
	svcInst  IService
	instance any
}

func (r *RuntimeAssembler) init(svcCtx service.Context, instance any) {
	r.svcInst = reinterpret.Cast[IService](svcCtx)
	r.instance = instance
}

func (r *RuntimeAssembler) assemble(settings _RuntimeSettings) core.Runtime {
	rtInstFace := iface.Face[runtime.Context]{}

	if cb, ok := r.instance.(IRuntimeInstantiator); ok {
		rtInstFace = iface.NewFaceTReflectC[runtime.Context, IRuntime](cb.Instantiate())
	} else {
		rtInstFace = iface.NewFaceTReflectC[runtime.Context, IRuntime](&RuntimeBehavior{})
	}

	rtInstFrameLoopBeginCB, _ := rtInstFace.Iface.(LifecycleRuntimeFrameLoopBegin)
	rtInstFrameUpdateBeginCB, _ := rtInstFace.Iface.(LifecycleRuntimeFrameUpdateBegin)
	rtInstFrameUpdateEndCB, _ := rtInstFace.Iface.(LifecycleRuntimeFrameUpdateEnd)
	rtInstFrameLoopEndCB, _ := rtInstFace.Iface.(LifecycleRuntimeFrameLoopEnd)
	rtInstRunCallBeginCB, _ := rtInstFace.Iface.(LifecycleRuntimeRunCallBegin)
	rtInstRunCallEndCB, _ := rtInstFace.Iface.(LifecycleRuntimeRunCallEnd)
	rtInstRunGCBeginCB, _ := rtInstFace.Iface.(LifecycleRuntimeRunGCBegin)
	rtInstRunGCEndCB, _ := rtInstFace.Iface.(LifecycleRuntimeRunGCEnd)

	frameLoopBeginCB, _ := r.instance.(LifecycleRuntimeFrameLoopBegin)
	frameUpdateBeginCB, _ := r.instance.(LifecycleRuntimeFrameUpdateBegin)
	frameUpdateEndCB, _ := r.instance.(LifecycleRuntimeFrameUpdateEnd)
	frameLoopEndCB, _ := r.instance.(LifecycleRuntimeFrameLoopEnd)
	runCallBeginCB, _ := r.instance.(LifecycleRuntimeRunCallBegin)
	runCallEndCB, _ := r.instance.(LifecycleRuntimeRunCallEnd)
	runGCBeginCB, _ := r.instance.(LifecycleRuntimeRunGCBegin)
	runGCEndCB, _ := r.instance.(LifecycleRuntimeRunGCEnd)

	rtCtx := runtime.NewContext(r.svcInst,
		runtime.With.InstanceFace(rtInstFace),
		runtime.With.Name(settings.name),
		runtime.With.PersistId(settings.persistId),
		runtime.With.PanicHandling(settings.autoRecover, settings.reportError),
		runtime.With.RunningEventCB(func(rtCtx runtime.Context, runningEvent runtime.RunningEvent, args ...any) {
			rtInst := reinterpret.Cast[IRuntime](rtCtx)

			switch runningEvent {
			case runtime.RunningEvent_Birth:
				cacheCallPath("", rtInst.Reflected().Type())
				rtInst.(iRuntime).setAutoInjection(settings.autoInjection)

				if cb, ok := r.instance.(LifecycleRuntimeBirth); ok {
					cb.Birth(rtInst)
				}
				if cb, ok := rtInst.(LifecycleRuntimeBirth); ok {
					cb.Birth(rtInst)
				}

				r.installAddIns(rtInst)

				if cb, ok := r.instance.(LifecycleRuntimeBuilt); ok {
					cb.Built(rtInst)
				}
				if cb, ok := rtInst.(LifecycleRuntimeBuilt); ok {
					cb.Built(rtInst)
				}
			case runtime.RunningEvent_Starting:
				if cb, ok := r.instance.(LifecycleRuntimeStarting); ok {
					cb.Starting(rtInst)
				}
				if cb, ok := rtInst.(LifecycleRuntimeStarting); ok {
					cb.Starting(rtInst)
				}
			case runtime.RunningEvent_Started:
				if cb, ok := r.instance.(LifecycleRuntimeStarted); ok {
					cb.Started(rtInst)
				}
				if cb, ok := rtInst.(LifecycleRuntimeStarted); ok {
					cb.Started(rtInst)
				}
			case runtime.RunningEvent_FrameLoopBegin:
				if cb := frameLoopBeginCB; cb != nil {
					cb.FrameLoopBegin(rtInst)
				}
				if cb := rtInstFrameLoopBeginCB; cb != nil {
					cb.FrameLoopBegin(rtInst)
				}
			case runtime.RunningEvent_FrameUpdateBegin:
				if cb := frameUpdateBeginCB; cb != nil {
					cb.FrameUpdateBegin(rtInst)
				}
				if cb := rtInstFrameUpdateBeginCB; cb != nil {
					cb.FrameUpdateBegin(rtInst)
				}
			case runtime.RunningEvent_FrameUpdateEnd:
				if cb := frameUpdateEndCB; cb != nil {
					cb.FrameUpdateEnd(rtInst)
				}
				if cb := rtInstFrameUpdateEndCB; cb != nil {
					cb.FrameUpdateEnd(rtInst)
				}
			case runtime.RunningEvent_FrameLoopEnd:
				if cb := frameLoopEndCB; cb != nil {
					cb.FrameLoopEnd(rtInst)
				}
				if cb := rtInstFrameLoopEndCB; cb != nil {
					cb.FrameLoopEnd(rtInst)
				}
			case runtime.RunningEvent_RunCallBegin:
				if cb := runCallBeginCB; cb != nil {
					cb.RunCallBegin(rtInst)
				}
				if cb := rtInstRunCallBeginCB; cb != nil {
					cb.RunCallBegin(rtInst)
				}
			case runtime.RunningEvent_RunCallEnd:
				if cb := runCallEndCB; cb != nil {
					cb.RunCallEnd(rtInst)
				}
				if cb := rtInstRunCallEndCB; cb != nil {
					cb.RunCallEnd(rtInst)
				}
			case runtime.RunningEvent_RunGCBegin:
				if cb := runGCBeginCB; cb != nil {
					cb.RunGCBegin(rtInst)
				}
				if cb := rtInstRunGCBeginCB; cb != nil {
					cb.RunGCBegin(rtInst)
				}
			case runtime.RunningEvent_RunGCEnd:
				if cb := runGCEndCB; cb != nil {
					cb.RunGCEnd(rtInst)
				}
				if cb := rtInstRunGCEndCB; cb != nil {
					cb.RunGCEnd(rtInst)
				}
			case runtime.RunningEvent_Terminating:
				if cb, ok := r.instance.(LifecycleRuntimeTerminating); ok {
					cb.Terminating(rtInst)
				}
				if cb, ok := rtInst.(LifecycleRuntimeTerminating); ok {
					cb.Terminating(rtInst)
				}
			case runtime.RunningEvent_Terminated:
				if cb, ok := r.instance.(LifecycleRuntimeTerminated); ok {
					cb.Terminated(rtInst)
				}
				if cb, ok := rtInst.(LifecycleRuntimeTerminated); ok {
					cb.Terminated(rtInst)
				}
			case runtime.RunningEvent_AddInActivating:
				addInStatus := args[0].(extension.AddInStatus)
				cacheCallPath(addInStatus.Name(), addInStatus.Reflected().Type())
				if cb, ok := r.instance.(LifecycleRuntimeAddInActivating); ok {
					cb.AddInActivating(rtInst, addInStatus)
				}
				if cb, ok := rtInst.(LifecycleRuntimeAddInActivating); ok {
					cb.AddInActivating(rtInst, addInStatus)
				}
			case runtime.RunningEvent_AddInActivated:
				addInStatus := args[0].(extension.AddInStatus)
				if cb, ok := r.instance.(LifecycleRuntimeAddInActivated); ok {
					cb.AddInActivated(rtInst, addInStatus)
				}
				if cb, ok := rtInst.(LifecycleRuntimeAddInActivated); ok {
					cb.AddInActivated(rtInst, addInStatus)
				}
			case runtime.RunningEvent_AddInDeactivating:
				addInStatus := args[0].(extension.AddInStatus)
				if cb, ok := r.instance.(LifecycleRuntimeAddInDeactivating); ok {
					cb.AddInDeactivating(rtInst, addInStatus)
				}
				if cb, ok := rtInst.(LifecycleRuntimeAddInDeactivating); ok {
					cb.AddInDeactivating(rtInst, addInStatus)
				}
			case runtime.RunningEvent_AddInDeactivated:
				addInStatus := args[0].(extension.AddInStatus)
				if cb, ok := r.instance.(LifecycleRuntimeAddInDeactivated); ok {
					cb.AddInDeactivated(rtInst, addInStatus)
				}
				if cb, ok := rtInst.(LifecycleRuntimeAddInDeactivated); ok {
					cb.AddInDeactivated(rtInst, addInStatus)
				}
			case runtime.RunningEvent_EntityActivating:
				entity := args[0].(ec.Entity)

				if entity.PT().Prototype() == "" {
					cacheCallPath("", entity.Reflected().Type())
				}

				if rtInst.AutoInjection() {
					ec.UnsafeEntity(entity).ComponentList().Traversal(func(compSlot *generic.FreeSlot[ec.Component]) bool {
						assertion.InjectRV(entity, compSlot.V.Reflected())
						return true
					})
				}
			case runtime.RunningEvent_EntityAddingComponents:
				entity := args[0].(ec.Entity)
				components := args[1].([]ec.Component)

				for i := range components {
					comp := components[i]
					cacheCallPath(comp.Name(), comp.Reflected().Type())
				}

				if rtInst.AutoInjection() {
					ec.UnsafeEntity(entity).ComponentList().Traversal(func(compSlot *generic.FreeSlot[ec.Component]) bool {
						assertion.InjectRV(entity, compSlot.V.Reflected())
						return true
					})
				}
			}
		}),
	)

	return core.NewRuntime(rtCtx,
		core.With.Runtime.AutoRun(true),
		core.With.Runtime.ContinueOnActivatingEntityPanic(settings.continueOnActivatingEntityPanic),
		core.With.Runtime.TaskQueue(core.With.TaskQueue.Unbounded(true)),
		core.With.Runtime.Frame(core.With.Frame.Enabled(settings.enableFrame), core.With.Frame.TargetFPS(settings.fps)),
	)
}

func (r *RuntimeAssembler) installAddIns(rtInst IRuntime) {
	appConf := r.svcInst.AppConf()

	installed := func(name string) bool {
		_, ok := rtInst.AddInManager().GetStatusByName(name)
		return ok
	}

	// 安装日志插件
	if !installed(Log.Name) {
		if cb, ok := rtInst.(InstallRuntimeLogger); ok {
			cb.InstallLogger(rtInst)
		}
	}
	if !installed(Log.Name) {
		if cb, ok := r.instance.(InstallRuntimeLogger); ok {
			cb.InstallLogger(rtInst)
		}
	}
	if !installed(Log.Name) {
		v, _ := r.svcInst.Memory().Load(memLogger)
		if logger, ok := v.(*zap.Logger); ok {
			Log.Install(rtInst,
				LogWith.Logger(logger),
			)
		}
	}

	// 安装RPC调用堆栈支持
	if !installed(RPCStack.Name) {
		if cb, ok := rtInst.(InstallRuntimeRPCStack); ok {
			cb.InstallRPCStack(rtInst)
		}
	}
	if !installed(RPCStack.Name) {
		if cb, ok := r.instance.(InstallRuntimeRPCStack); ok {
			cb.InstallRPCStack(rtInst)
		}
	}
	if !installed(RPCStack.Name) {
		RPCStack.Install(rtInst)
	}

	// 安装分布式实体支持插件
	if !installed(Dentr.Name) {
		if cb, ok := rtInst.(InstallRuntimeDistEntityRegistry); ok {
			cb.InstallDistEntityRegistry(rtInst)
		}
	}
	if !installed(Dentr.Name) {
		if cb, ok := r.instance.(InstallRuntimeDistEntityRegistry); ok {
			cb.InstallDistEntityRegistry(rtInst)
		}
	}
	if !installed(Dentr.Name) {
		v, _ := r.svcInst.Memory().Load(memEtcdClientOnce)
		cliOnce, ok := v.(func() *etcdv3.Client)
		if !ok {
			exception.Panicf("%w: service memory %q not exists", ErrFramework, memEtcdClientOnce)
		}
		Dentr.Install(rtInst,
			DentrWith.EtcdClient(cliOnce()),
			DentrWith.RegistrationTTL(appConf.GetDuration("service.dent_ttl")),
		)
	}
}
