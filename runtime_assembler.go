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
	"fmt"

	"git.golaxy.org/core"
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/extension"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/utils/assertion"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/iface"
	"git.golaxy.org/core/utils/reinterpret"
	"git.golaxy.org/core/utils/uid"
	. "git.golaxy.org/framework/addins"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

type _RuntimeSettings struct {
	name                            string
	persistId                       uid.Id
	mainEntity                      ec.Entity
	autoRecover                     bool
	reportError                     chan error
	continueOnActivatingEntityPanic bool
	enableFrame                     bool
	fps                             float64
	autoInjection                   bool
}

type iRuntimeAssembler interface {
	init(svcInst IService, instance any)
	assemble(settings _RuntimeSettings) (core.Runtime, error)
}

// RuntimeAssembler 运行时装配器
type RuntimeAssembler struct {
	svcInst  IService
	instance any
}

func (r *RuntimeAssembler) init(svcInst IService, instance any) {
	r.svcInst = svcInst
	r.instance = instance
}

func (r *RuntimeAssembler) assemble(settings _RuntimeSettings) (core.Runtime, error) {
	rtInstFace := iface.Face[runtime.Context]{}

	if cb, ok := r.instance.(IRuntimeInstantiator); ok {
		rtInstFace = iface.NewFaceTReflectC[runtime.Context, IRuntime](cb.Instantiate())
	} else {
		rtInstFace = iface.NewFaceTReflectC[runtime.Context, IRuntime](&RuntimeBehavior{})
	}

	iRtInst := rtInstFace.Iface.(iRuntime)
	iRtInst.setMainEntity(settings.mainEntity)
	iRtInst.setAutoInjection(settings.autoInjection)

	rtInstFrameLoopBeginCB, _ := rtInstFace.Iface.(LifecycleRuntimeFrameLoopBegin)
	rtInstFrameUpdateBeginCB, _ := rtInstFace.Iface.(LifecycleRuntimeFrameUpdateBegin)
	rtInstFrameUpdateEndCB, _ := rtInstFace.Iface.(LifecycleRuntimeFrameUpdateEnd)
	rtInstFrameLoopEndCB, _ := rtInstFace.Iface.(LifecycleRuntimeFrameLoopEnd)
	rtInstRunCallBeginCB, _ := rtInstFace.Iface.(LifecycleRuntimeRunCallBegin)
	rtInstRunCallEndCB, _ := rtInstFace.Iface.(LifecycleRuntimeRunCallEnd)
	rtInstRunGCBeginCB, _ := rtInstFace.Iface.(LifecycleRuntimeRunGCBegin)
	rtInstRunGCEndCB, _ := rtInstFace.Iface.(LifecycleRuntimeRunGCEnd)
	rtInstEntityActivatingCB, _ := rtInstFace.Iface.(LifecycleRuntimeEntityActivating)
	rtInstEntityActivationAbortedCB, _ := rtInstFace.Iface.(LifecycleRuntimeEntityActivationAborted)
	rtInstEntityActivatedCB, _ := rtInstFace.Iface.(LifecycleRuntimeEntityActivated)
	rtInstEntityDeactivatingCB, _ := rtInstFace.Iface.(LifecycleRuntimeEntityDeactivating)
	rtInstEntityDeactivatedCB, _ := rtInstFace.Iface.(LifecycleRuntimeEntityDeactivated)
	rtInstEntityAddingComponentsCB, _ := rtInstFace.Iface.(LifecycleRuntimeEntityAddingComponents)
	rtInstEntityComponentsAdditionAbortedCB, _ := rtInstFace.Iface.(LifecycleRuntimeEntityComponentsAdditionAborted)
	rtInstEntityComponentsAddedCB, _ := rtInstFace.Iface.(LifecycleRuntimeEntityComponentsAdded)
	rtInstEntityRemovingComponentCB, _ := rtInstFace.Iface.(LifecycleRuntimeEntityRemovingComponent)
	rtInstEntityComponentRemovalAbortedCB, _ := rtInstFace.Iface.(LifecycleRuntimeEntityComponentRemovalAborted)
	rtInstEntityComponentRemovedCB, _ := rtInstFace.Iface.(LifecycleRuntimeEntityComponentRemoved)

	frameLoopBeginCB, _ := r.instance.(LifecycleRuntimeFrameLoopBegin)
	frameUpdateBeginCB, _ := r.instance.(LifecycleRuntimeFrameUpdateBegin)
	frameUpdateEndCB, _ := r.instance.(LifecycleRuntimeFrameUpdateEnd)
	frameLoopEndCB, _ := r.instance.(LifecycleRuntimeFrameLoopEnd)
	runCallBeginCB, _ := r.instance.(LifecycleRuntimeRunCallBegin)
	runCallEndCB, _ := r.instance.(LifecycleRuntimeRunCallEnd)
	runGCBeginCB, _ := r.instance.(LifecycleRuntimeRunGCBegin)
	runGCEndCB, _ := r.instance.(LifecycleRuntimeRunGCEnd)
	entityActivatingCB, _ := r.instance.(LifecycleRuntimeEntityActivating)
	entityActivationAbortedCB, _ := r.instance.(LifecycleRuntimeEntityActivationAborted)
	entityActivatedCB, _ := r.instance.(LifecycleRuntimeEntityActivated)
	entityDeactivatingCB, _ := r.instance.(LifecycleRuntimeEntityDeactivating)
	entityDeactivatedCB, _ := r.instance.(LifecycleRuntimeEntityDeactivated)
	entityAddingComponentsCB, _ := r.instance.(LifecycleRuntimeEntityAddingComponents)
	entityComponentsAdditionAbortedCB, _ := r.instance.(LifecycleRuntimeEntityComponentsAdditionAborted)
	entityComponentsAddedCB, _ := r.instance.(LifecycleRuntimeEntityComponentsAdded)
	entityRemovingComponentCB, _ := r.instance.(LifecycleRuntimeEntityRemovingComponent)
	entityComponentRemovalAbortedCB, _ := r.instance.(LifecycleRuntimeEntityComponentRemovalAborted)
	entityComponentRemovedCB, _ := r.instance.(LifecycleRuntimeEntityComponentRemoved)

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

				if cb, ok := r.instance.(LifecycleRuntimeBirth); ok {
					cb.OnBirth(rtInst)
				}
				if cb, ok := rtInst.(LifecycleRuntimeBirth); ok {
					cb.OnBirth(rtInst)
				}

				r.installAddIns(rtInst)

				if cb, ok := r.instance.(LifecycleRuntimeBuilt); ok {
					cb.OnBuilt(rtInst)
				}
				if cb, ok := rtInst.(LifecycleRuntimeBuilt); ok {
					cb.OnBuilt(rtInst)
				}
			case runtime.RunningEvent_Starting:
				if cb, ok := r.instance.(LifecycleRuntimeStarting); ok {
					cb.OnStarting(rtInst)
				}
				if cb, ok := rtInst.(LifecycleRuntimeStarting); ok {
					cb.OnStarting(rtInst)
				}
			case runtime.RunningEvent_Started:
				if cb, ok := r.instance.(LifecycleRuntimeStarted); ok {
					cb.OnStarted(rtInst)
				}
				if cb, ok := rtInst.(LifecycleRuntimeStarted); ok {
					cb.OnStarted(rtInst)
				}
			case runtime.RunningEvent_FrameLoopBegin:
				if cb := frameLoopBeginCB; cb != nil {
					cb.OnFrameLoopBegin(rtInst)
				}
				if cb := rtInstFrameLoopBeginCB; cb != nil {
					cb.OnFrameLoopBegin(rtInst)
				}
			case runtime.RunningEvent_FrameUpdateBegin:
				if cb := frameUpdateBeginCB; cb != nil {
					cb.OnFrameUpdateBegin(rtInst)
				}
				if cb := rtInstFrameUpdateBeginCB; cb != nil {
					cb.OnFrameUpdateBegin(rtInst)
				}
			case runtime.RunningEvent_FrameUpdateEnd:
				if cb := frameUpdateEndCB; cb != nil {
					cb.OnFrameUpdateEnd(rtInst)
				}
				if cb := rtInstFrameUpdateEndCB; cb != nil {
					cb.OnFrameUpdateEnd(rtInst)
				}
			case runtime.RunningEvent_FrameLoopEnd:
				if cb := frameLoopEndCB; cb != nil {
					cb.OnFrameLoopEnd(rtInst)
				}
				if cb := rtInstFrameLoopEndCB; cb != nil {
					cb.OnFrameLoopEnd(rtInst)
				}
			case runtime.RunningEvent_RunCallBegin:
				if cb := runCallBeginCB; cb != nil {
					cb.OnRunCallBegin(rtInst)
				}
				if cb := rtInstRunCallBeginCB; cb != nil {
					cb.OnRunCallBegin(rtInst)
				}
			case runtime.RunningEvent_RunCallEnd:
				if cb := runCallEndCB; cb != nil {
					cb.OnRunCallEnd(rtInst)
				}
				if cb := rtInstRunCallEndCB; cb != nil {
					cb.OnRunCallEnd(rtInst)
				}
			case runtime.RunningEvent_RunGCBegin:
				if cb := runGCBeginCB; cb != nil {
					cb.OnRunGCBegin(rtInst)
				}
				if cb := rtInstRunGCBeginCB; cb != nil {
					cb.OnRunGCBegin(rtInst)
				}
			case runtime.RunningEvent_RunGCEnd:
				if cb := runGCEndCB; cb != nil {
					cb.OnRunGCEnd(rtInst)
				}
				if cb := rtInstRunGCEndCB; cb != nil {
					cb.OnRunGCEnd(rtInst)
				}
			case runtime.RunningEvent_Terminating:
				if cb, ok := r.instance.(LifecycleRuntimeTerminating); ok {
					cb.OnTerminating(rtInst)
				}
				if cb, ok := rtInst.(LifecycleRuntimeTerminating); ok {
					cb.OnTerminating(rtInst)
				}
			case runtime.RunningEvent_Terminated:
				if cb, ok := r.instance.(LifecycleRuntimeTerminated); ok {
					cb.OnTerminated(rtInst)
				}
				if cb, ok := rtInst.(LifecycleRuntimeTerminated); ok {
					cb.OnTerminated(rtInst)
				}
			case runtime.RunningEvent_AddInActivating:
				addInStatus := args[0].(extension.AddInStatus)
				cacheCallPath(addInStatus.Name(), addInStatus.Reflected().Type())
				if cb, ok := r.instance.(LifecycleRuntimeAddInActivating); ok {
					cb.OnAddInActivating(rtInst, addInStatus)
				}
				if cb, ok := rtInst.(LifecycleRuntimeAddInActivating); ok {
					cb.OnAddInActivating(rtInst, addInStatus)
				}
			case runtime.RunningEvent_AddInActivationAborted:
				addInStatus := args[0].(extension.AddInStatus)
				if cb, ok := r.instance.(LifecycleRuntimeAddInActivationAborted); ok {
					cb.OnAddInActivationAborted(rtInst, addInStatus)
				}
				if cb, ok := rtInst.(LifecycleRuntimeAddInActivationAborted); ok {
					cb.OnAddInActivationAborted(rtInst, addInStatus)
				}
			case runtime.RunningEvent_AddInActivated:
				addInStatus := args[0].(extension.AddInStatus)
				if cb, ok := r.instance.(LifecycleRuntimeAddInActivated); ok {
					cb.OnAddInActivated(rtInst, addInStatus)
				}
				if cb, ok := rtInst.(LifecycleRuntimeAddInActivated); ok {
					cb.OnAddInActivated(rtInst, addInStatus)
				}
			case runtime.RunningEvent_AddInDeactivating:
				addInStatus := args[0].(extension.AddInStatus)
				if cb, ok := r.instance.(LifecycleRuntimeAddInDeactivating); ok {
					cb.OnAddInDeactivating(rtInst, addInStatus)
				}
				if cb, ok := rtInst.(LifecycleRuntimeAddInDeactivating); ok {
					cb.OnAddInDeactivating(rtInst, addInStatus)
				}
			case runtime.RunningEvent_AddInDeactivated:
				addInStatus := args[0].(extension.AddInStatus)
				if cb, ok := r.instance.(LifecycleRuntimeAddInDeactivated); ok {
					cb.OnAddInDeactivated(rtInst, addInStatus)
				}
				if cb, ok := rtInst.(LifecycleRuntimeAddInDeactivated); ok {
					cb.OnAddInDeactivated(rtInst, addInStatus)
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

				if cb := entityActivatingCB; cb != nil {
					cb.OnEntityActivating(rtInst, entity)
				}
				if cb := rtInstEntityActivatingCB; cb != nil {
					cb.OnEntityActivating(rtInst, entity)
				}
			case runtime.RunningEvent_EntityActivationAborted:
				entity := args[0].(ec.Entity)
				if cb := entityActivationAbortedCB; cb != nil {
					cb.OnEntityActivationAborted(rtInst, entity)
				}
				if cb := rtInstEntityActivationAbortedCB; cb != nil {
					cb.OnEntityActivationAborted(rtInst, entity)
				}
			case runtime.RunningEvent_EntityActivated:
				entity := args[0].(ec.Entity)
				if cb := entityActivatedCB; cb != nil {
					cb.OnEntityActivated(rtInst, entity)
				}
				if cb := rtInstEntityActivatedCB; cb != nil {
					cb.OnEntityActivated(rtInst, entity)
				}
			case runtime.RunningEvent_EntityDeactivating:
				entity := args[0].(ec.Entity)
				if cb := entityDeactivatingCB; cb != nil {
					cb.OnEntityDeactivating(rtInst, entity)
				}
				if cb := rtInstEntityDeactivatingCB; cb != nil {
					cb.OnEntityDeactivating(rtInst, entity)
				}
			case runtime.RunningEvent_EntityDeactivated:
				entity := args[0].(ec.Entity)
				if cb := entityDeactivatedCB; cb != nil {
					cb.OnEntityDeactivated(rtInst, entity)
				}
				if cb := rtInstEntityDeactivatedCB; cb != nil {
					cb.OnEntityDeactivated(rtInst, entity)
				}
				if settings.mainEntity == entity {
					rtCtx.Terminate()
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

				if cb := entityAddingComponentsCB; cb != nil {
					cb.OnEntityAddingComponents(rtInst, entity, components)
				}
				if cb := rtInstEntityAddingComponentsCB; cb != nil {
					cb.OnEntityAddingComponents(rtInst, entity, components)
				}
			case runtime.RunningEvent_EntityComponentsAdditionAborted:
				entity := args[0].(ec.Entity)
				components := args[1].([]ec.Component)
				if cb := entityComponentsAdditionAbortedCB; cb != nil {
					cb.OnEntityComponentsAdditionAborted(rtInst, entity, components)
				}
				if cb := rtInstEntityComponentsAdditionAbortedCB; cb != nil {
					cb.OnEntityComponentsAdditionAborted(rtInst, entity, components)
				}
			case runtime.RunningEvent_EntityComponentsAdded:
				entity := args[0].(ec.Entity)
				components := args[1].([]ec.Component)
				if cb := entityComponentsAddedCB; cb != nil {
					cb.OnEntityComponentsAdded(rtInst, entity, components)
				}
				if cb := rtInstEntityComponentsAddedCB; cb != nil {
					cb.OnEntityComponentsAdded(rtInst, entity, components)
				}
			case runtime.RunningEvent_EntityRemovingComponent:
				entity := args[0].(ec.Entity)
				component := args[1].(ec.Component)
				if cb := entityRemovingComponentCB; cb != nil {
					cb.OnEntityRemovingComponent(rtInst, entity, component)
				}
				if cb := rtInstEntityRemovingComponentCB; cb != nil {
					cb.OnEntityRemovingComponent(rtInst, entity, component)
				}
			case runtime.RunningEvent_EntityComponentRemovalAborted:
				entity := args[0].(ec.Entity)
				component := args[1].(ec.Component)
				if cb := entityComponentRemovalAbortedCB; cb != nil {
					cb.OnEntityComponentRemovalAborted(rtInst, entity, component)
				}
				if cb := rtInstEntityComponentRemovalAbortedCB; cb != nil {
					cb.OnEntityComponentRemovalAborted(rtInst, entity, component)
				}
			case runtime.RunningEvent_EntityComponentRemoved:
				entity := args[0].(ec.Entity)
				component := args[1].(ec.Component)
				if cb := entityComponentRemovedCB; cb != nil {
					cb.OnEntityComponentRemoved(rtInst, entity, component)
				}
				if cb := rtInstEntityComponentRemovedCB; cb != nil {
					cb.OnEntityComponentRemoved(rtInst, entity, component)
				}
			}
		}),
	)

	if settings.mainEntity != nil {
		if err := rtCtx.EntityManager().AddEntity(settings.mainEntity); err != nil {
			return nil, fmt.Errorf("%w: failed to add main entity, %w", ErrFramework, err)
		}
	}

	return core.NewRuntime(rtCtx,
		core.With.Runtime.AutoRun(true),
		core.With.Runtime.ContinueOnActivatingEntityPanic(settings.continueOnActivatingEntityPanic),
		core.With.Runtime.TaskQueue(core.With.TaskQueue.Unbounded(true)),
		core.With.Runtime.Frame(core.With.Frame.Enabled(settings.enableFrame), core.With.Frame.TargetFPS(settings.fps)),
	), nil
}

func (r *RuntimeAssembler) installAddIns(rtInst IRuntime) {
	conf := r.svcInst.AppConf()

	installed := func(name string) bool {
		_, ok := rtInst.AddInManager().GetStatusByName(name)
		return ok
	}
	requireInstalled := func(name string) {
		if !installed(name) {
			exception.Panicf("%w: runtime add-in %q not installed", ErrFramework, name)
		}
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
		Log.Install(rtInst,
			LogWith.Logger(r.svcInst.L()),
		)
	}
	requireInstalled(Log.Name)

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
	requireInstalled(RPCStack.Name)

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
			DentrWith.RegistrationTTL(conf.GetDuration("service.dent_ttl")),
		)
	}
	requireInstalled(Dentr.Name)
}
