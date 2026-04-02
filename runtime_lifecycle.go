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

type LifecycleRuntimeBuilt interface {
	OnBuilt(rt IRuntime)
}

type LifecycleRuntimeBirth interface {
	OnBirth(rt IRuntime)
}

type LifecycleRuntimeStarting interface {
	OnStarting(rt IRuntime)
}

type LifecycleRuntimeStarted interface {
	OnStarted(rt IRuntime)
}

type LifecycleRuntimeFrameLoopBegin interface {
	OnFrameLoopBegin(rt IRuntime)
}

type LifecycleRuntimeFrameUpdateBegin interface {
	OnFrameUpdateBegin(rt IRuntime)
}

type LifecycleRuntimeFrameUpdateEnd interface {
	OnFrameUpdateEnd(rt IRuntime)
}

type LifecycleRuntimeFrameLoopEnd interface {
	OnFrameLoopEnd(rt IRuntime)
}

type LifecycleRuntimeRunCallBegin interface {
	OnRunCallBegin(rt IRuntime)
}

type LifecycleRuntimeRunCallEnd interface {
	OnRunCallEnd(rt IRuntime)
}

type LifecycleRuntimeRunGCBegin interface {
	OnRunGCBegin(rt IRuntime)
}

type LifecycleRuntimeRunGCEnd interface {
	OnRunGCEnd(rt IRuntime)
}

type LifecycleRuntimeTerminating interface {
	OnTerminating(rt IRuntime)
}

type LifecycleRuntimeTerminated interface {
	OnTerminated(rt IRuntime)
}

type LifecycleRuntimeAddInActivating interface {
	OnAddInActivating(rt IRuntime, addIn extension.AddInStatus)
}

type LifecycleRuntimeAddInActivated interface {
	OnAddInActivated(rt IRuntime, addIn extension.AddInStatus)
}

type LifecycleRuntimeAddInDeactivating interface {
	OnAddInDeactivating(rt IRuntime, addIn extension.AddInStatus)
}

type LifecycleRuntimeAddInDeactivated interface {
	OnAddInDeactivated(rt IRuntime, addIn extension.AddInStatus)
}

type LifecycleRuntimeEntityActivating interface {
	OnEntityActivating(rt IRuntime, entity ec.Entity)
}

type LifecycleRuntimeEntityActivated interface {
	OnEntityActivated(rt IRuntime, entity ec.Entity)
}

type LifecycleRuntimeEntityAddingComponents interface {
	OnEntityAddingComponents(rt IRuntime, entity ec.Entity, components []ec.Component)
}

type LifecycleRuntimeEntityComponentsAdded interface {
	OnEntityComponentsAdded(rt IRuntime, entity ec.Entity, components []ec.Component)
}

type LifecycleRuntimeEntityRemovingComponent interface {
	OnEntityRemovingComponent(rt IRuntime, entity ec.Entity, component ec.Component)
}

type LifecycleRuntimeEntityComponentRemoved interface {
	OnEntityComponentRemoved(rt IRuntime, entity ec.Entity, component ec.Component)
}
