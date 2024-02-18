package framework

import "git.golaxy.org/core/runtime"

type LifecycleRuntimeBirth interface {
	Birth(ctx runtime.Context)
}

type LifecycleRuntimeInit interface {
	Init(ctx runtime.Context)
}

type LifecycleRuntimeStarting interface {
	Starting(ctx runtime.Context)
}

type LifecycleRuntimeStarted interface {
	Started(ctx runtime.Context)
}

type LifecycleRuntimeFrameLoopBegin interface {
	FrameLoopBegin(ctx runtime.Context)
}

type LifecycleRuntimeFrameUpdateBegin interface {
	FrameUpdateBegin(ctx runtime.Context)
}

type LifecycleRuntimeFrameUpdateEnd interface {
	FrameUpdateEnd(ctx runtime.Context)
}

type LifecycleRuntimeFrameLoopEnd interface {
	FrameLoopEnd(ctx runtime.Context)
}

type LifecycleRuntimeRunCallBegin interface {
	RunCallBegin(ctx runtime.Context)
}

type LifecycleRuntimeRunCallEnd interface {
	RunCallEnd(ctx runtime.Context)
}

type LifecycleRuntimeRunGCBegin interface {
	RunGCBegin(ctx runtime.Context)
}

type LifecycleRuntimeRunGCEnd interface {
	RunGCEnd(ctx runtime.Context)
}

type LifecycleRuntimeTerminating interface {
	Terminating(ctx runtime.Context)
}

type LifecycleRuntimeTerminated interface {
	Terminated(ctx runtime.Context)
}
