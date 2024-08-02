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

type LifecycleRuntimeBuilt interface {
	Built(inst IRuntimeInstance)
}

type LifecycleRuntimeBirth interface {
	Birth(inst IRuntimeInstance)
}

type LifecycleRuntimeStarting interface {
	Starting(inst IRuntimeInstance)
}

type LifecycleRuntimeStarted interface {
	Started(inst IRuntimeInstance)
}

type LifecycleRuntimeFrameLoopBegin interface {
	FrameLoopBegin(inst IRuntimeInstance)
}

type LifecycleRuntimeFrameUpdateBegin interface {
	FrameUpdateBegin(inst IRuntimeInstance)
}

type LifecycleRuntimeFrameUpdateEnd interface {
	FrameUpdateEnd(inst IRuntimeInstance)
}

type LifecycleRuntimeFrameLoopEnd interface {
	FrameLoopEnd(inst IRuntimeInstance)
}

type LifecycleRuntimeRunCallBegin interface {
	RunCallBegin(inst IRuntimeInstance)
}

type LifecycleRuntimeRunCallEnd interface {
	RunCallEnd(inst IRuntimeInstance)
}

type LifecycleRuntimeRunGCBegin interface {
	RunGCBegin(inst IRuntimeInstance)
}

type LifecycleRuntimeRunGCEnd interface {
	RunGCEnd(inst IRuntimeInstance)
}

type LifecycleRuntimeTerminating interface {
	Terminating(inst IRuntimeInstance)
}

type LifecycleRuntimeTerminated interface {
	Terminated(inst IRuntimeInstance)
}
