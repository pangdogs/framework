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
	Built(inst IRuntime)
}

type LifecycleRuntimeBirth interface {
	Birth(inst IRuntime)
}

type LifecycleRuntimeStarting interface {
	Starting(inst IRuntime)
}

type LifecycleRuntimeStarted interface {
	Started(inst IRuntime)
}

type LifecycleRuntimeFrameLoopBegin interface {
	FrameLoopBegin(inst IRuntime)
}

type LifecycleRuntimeFrameUpdateBegin interface {
	FrameUpdateBegin(inst IRuntime)
}

type LifecycleRuntimeFrameUpdateEnd interface {
	FrameUpdateEnd(inst IRuntime)
}

type LifecycleRuntimeFrameLoopEnd interface {
	FrameLoopEnd(inst IRuntime)
}

type LifecycleRuntimeRunCallBegin interface {
	RunCallBegin(inst IRuntime)
}

type LifecycleRuntimeRunCallEnd interface {
	RunCallEnd(inst IRuntime)
}

type LifecycleRuntimeRunGCBegin interface {
	RunGCBegin(inst IRuntime)
}

type LifecycleRuntimeRunGCEnd interface {
	RunGCEnd(inst IRuntime)
}

type LifecycleRuntimeTerminating interface {
	Terminating(inst IRuntime)
}

type LifecycleRuntimeTerminated interface {
	Terminated(inst IRuntime)
}

type LifecycleRuntimeAddInActivating interface {
	AddInActivating(inst IRuntime, addIn extension.AddInStatus)
}

type LifecycleRuntimeAddInActivated interface {
	AddInActivated(inst IRuntime, addIn extension.AddInStatus)
}

type LifecycleRuntimeAddInDeactivating interface {
	AddInDeactivating(inst IRuntime, addIn extension.AddInStatus)
}

type LifecycleRuntimeAddInDeactivated interface {
	AddInDeactivated(inst IRuntime, addIn extension.AddInStatus)
}
