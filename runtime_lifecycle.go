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

import "git.golaxy.org/core/extension"

type LifecycleRuntimeBuilt interface {
	Built(rt IRuntime)
}

type LifecycleRuntimeBirth interface {
	Birth(rt IRuntime)
}

type LifecycleRuntimeStarting interface {
	Starting(rt IRuntime)
}

type LifecycleRuntimeStarted interface {
	Started(rt IRuntime)
}

type LifecycleRuntimeFrameLoopBegin interface {
	FrameLoopBegin(rt IRuntime)
}

type LifecycleRuntimeFrameUpdateBegin interface {
	FrameUpdateBegin(rt IRuntime)
}

type LifecycleRuntimeFrameUpdateEnd interface {
	FrameUpdateEnd(rt IRuntime)
}

type LifecycleRuntimeFrameLoopEnd interface {
	FrameLoopEnd(rt IRuntime)
}

type LifecycleRuntimeRunCallBegin interface {
	RunCallBegin(rt IRuntime)
}

type LifecycleRuntimeRunCallEnd interface {
	RunCallEnd(rt IRuntime)
}

type LifecycleRuntimeRunGCBegin interface {
	RunGCBegin(rt IRuntime)
}

type LifecycleRuntimeRunGCEnd interface {
	RunGCEnd(rt IRuntime)
}

type LifecycleRuntimeTerminating interface {
	Terminating(rt IRuntime)
}

type LifecycleRuntimeTerminated interface {
	Terminated(rt IRuntime)
}

type LifecycleRuntimeAddInActivating interface {
	AddInActivating(rt IRuntime, addIn extension.AddInStatus)
}

type LifecycleRuntimeAddInActivated interface {
	AddInActivated(rt IRuntime, addIn extension.AddInStatus)
}

type LifecycleRuntimeAddInDeactivating interface {
	AddInDeactivating(rt IRuntime, addIn extension.AddInStatus)
}

type LifecycleRuntimeAddInDeactivated interface {
	AddInDeactivated(rt IRuntime, addIn extension.AddInStatus)
}
