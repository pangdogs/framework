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
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/extension"
)

// LifecycleRuntimeBuilt 在运行时完成 framework add-in 装配后接收回调。
// 此时运行时尚未进入 Starting，仍可补充初始状态。
type LifecycleRuntimeBuilt interface {
	OnBuilt(rt IRuntime)
}

// LifecycleRuntimeBirth 在运行时上下文创建后、framework add-in 装配前接收回调。
type LifecycleRuntimeBirth interface {
	OnBirth(rt IRuntime)
}

// LifecycleRuntimeStarting 在运行时工作循环启动时接收回调。
// 预装 add-in 已在回调前激活。
type LifecycleRuntimeStarting interface {
	OnStarting(rt IRuntime)
}

// LifecycleRuntimeStarted 在运行时工作循环及实体事件监听就绪后接收回调。
type LifecycleRuntimeStarted interface {
	OnStarted(rt IRuntime)
}

// LifecycleRuntimeFrameLoopBegin 在一次完整帧循环开始时接收回调。
type LifecycleRuntimeFrameLoopBegin interface {
	OnFrameLoopBegin(rt IRuntime)
}

// LifecycleRuntimeFrameUpdateBegin 在一次帧更新开始时接收回调。
type LifecycleRuntimeFrameUpdateBegin interface {
	OnFrameUpdateBegin(rt IRuntime)
}

// LifecycleRuntimeFrameUpdateEnd 在一次帧更新结束时接收回调。
type LifecycleRuntimeFrameUpdateEnd interface {
	OnFrameUpdateEnd(rt IRuntime)
}

// LifecycleRuntimeFrameLoopEnd 在一次完整帧循环结束时接收回调。
type LifecycleRuntimeFrameLoopEnd interface {
	OnFrameLoopEnd(rt IRuntime)
}

// LifecycleRuntimeRunCallBegin 在运行时开始执行一个普通异步调用任务时接收回调。
type LifecycleRuntimeRunCallBegin interface {
	OnRunCallBegin(rt IRuntime)
}

// LifecycleRuntimeRunCallEnd 在运行时执行完一个普通异步调用任务时接收回调。
type LifecycleRuntimeRunCallEnd interface {
	OnRunCallEnd(rt IRuntime)
}

// LifecycleRuntimeRunGCBegin 在运行时开始执行一次 GC 时接收回调。
type LifecycleRuntimeRunGCBegin interface {
	OnRunGCBegin(rt IRuntime)
}

// LifecycleRuntimeRunGCEnd 在运行时执行完一次 GC 时接收回调。
type LifecycleRuntimeRunGCEnd interface {
	OnRunGCEnd(rt IRuntime)
}

// LifecycleRuntimeTerminating 在运行时开始停止、等待子任务退出前接收回调。
type LifecycleRuntimeTerminating interface {
	OnTerminating(rt IRuntime)
}

// LifecycleRuntimeTerminated 在运行时等待组清空且 add-in 停用后接收回调。
type LifecycleRuntimeTerminated interface {
	OnTerminated(rt IRuntime)
}

// LifecycleRuntimeAddInActivating 在 add-in 调用 Init 前接收回调。
// 回调中卸载该 add-in 会中止激活流程。
type LifecycleRuntimeAddInActivating interface {
	OnAddInActivating(rt IRuntime, addIn extension.AddInStatus)
}

// LifecycleRuntimeAddInActivationAborted 在 add-in 激活流程被中止时接收回调。
type LifecycleRuntimeAddInActivationAborted interface {
	OnAddInActivationAborted(rt IRuntime, addIn extension.AddInStatus)
}

// LifecycleRuntimeAddInActivated 在 add-in 完成 Init 并进入运行状态后接收回调。
type LifecycleRuntimeAddInActivated interface {
	OnAddInActivated(rt IRuntime, addIn extension.AddInStatus)
}

// LifecycleRuntimeAddInDeactivating 在 add-in 调用 Shut 前接收回调。
type LifecycleRuntimeAddInDeactivating interface {
	OnAddInDeactivating(rt IRuntime, addIn extension.AddInStatus)
}

// LifecycleRuntimeAddInDeactivated 在 add-in 完成 Shut 后接收回调。
type LifecycleRuntimeAddInDeactivated interface {
	OnAddInDeactivated(rt IRuntime, addIn extension.AddInStatus)
}

// LifecycleRuntimeEntityActivating 在实体开始激活时接收回调。
type LifecycleRuntimeEntityActivating interface {
	OnEntityActivating(rt IRuntime, entity ec.Entity)
}

// LifecycleRuntimeEntityActivationAborted 在实体激活流程被中止时接收回调。
type LifecycleRuntimeEntityActivationAborted interface {
	OnEntityActivationAborted(rt IRuntime, entity ec.Entity)
}

// LifecycleRuntimeEntityActivated 在实体及其初始组件激活完成后接收回调。
type LifecycleRuntimeEntityActivated interface {
	OnEntityActivated(rt IRuntime, entity ec.Entity)
}

// LifecycleRuntimeEntityDeactivating 在实体开始停用时接收回调。
type LifecycleRuntimeEntityDeactivating interface {
	OnEntityDeactivating(rt IRuntime, entity ec.Entity)
}

// LifecycleRuntimeEntityDeactivated 在实体停用完成时接收回调。
type LifecycleRuntimeEntityDeactivated interface {
	OnEntityDeactivated(rt IRuntime, entity ec.Entity)
}

// LifecycleRuntimeEntityComponentsActivating 在一批新增组件开始激活时接收回调。
type LifecycleRuntimeEntityComponentsActivating interface {
	OnEntityComponentsActivating(rt IRuntime, entity ec.Entity, components []ec.Component)
}

// LifecycleRuntimeEntityComponentsActivationAborted 在新增组件的激活流程被中止时接收回调。
type LifecycleRuntimeEntityComponentsActivationAborted interface {
	OnEntityComponentsActivationAborted(rt IRuntime, entity ec.Entity, components []ec.Component)
}

// LifecycleRuntimeEntityComponentsActivated 在一批新增组件激活完成后接收回调。
type LifecycleRuntimeEntityComponentsActivated interface {
	OnEntityComponentsActivated(rt IRuntime, entity ec.Entity, components []ec.Component)
}

// LifecycleRuntimeEntityComponentDeactivating 在即将删除的组件开始停用时接收回调。
type LifecycleRuntimeEntityComponentDeactivating interface {
	OnEntityComponentDeactivating(rt IRuntime, entity ec.Entity, component ec.Component)
}

// LifecycleRuntimeEntityComponentDeactivationAborted 在组件停用回调流程被中止时接收回调。
// 该事件不会取消组件删除。
type LifecycleRuntimeEntityComponentDeactivationAborted interface {
	OnEntityComponentDeactivationAborted(rt IRuntime, entity ec.Entity, component ec.Component)
}

// LifecycleRuntimeEntityComponentDeactivated 在组件停用完成、即将从实体删除时接收回调。
type LifecycleRuntimeEntityComponentDeactivated interface {
	OnEntityComponentDeactivated(rt IRuntime, entity ec.Entity, component ec.Component)
}
