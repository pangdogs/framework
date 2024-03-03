package framework

import "git.golaxy.org/core/runtime"

type SetupRuntimeContextComposite interface {
	MakeContextComposite() runtime.Context
}

type LifecycleRuntimeContextBirth interface {
	Birth()
}

type LifecycleRuntimeContextInit interface {
	Init()
}

type LifecycleRuntimeContextStarting interface {
	Starting()
}

type LifecycleRuntimeContextStarted interface {
	Started()
}

type LifecycleRuntimeContextFrameLoopBegin interface {
	FrameLoopBegin()
}

type LifecycleRuntimeContextFrameUpdateBegin interface {
	FrameUpdateBegin()
}

type LifecycleRuntimeContextFrameUpdateEnd interface {
	FrameUpdateEnd()
}

type LifecycleRuntimeContextFrameLoopEnd interface {
	FrameLoopEnd()
}

type LifecycleRuntimeContextRunCallBegin interface {
	RunCallBegin()
}

type LifecycleRuntimeContextRunCallEnd interface {
	RunCallEnd()
}

type LifecycleRuntimeContextRunGCBegin interface {
	RunGCBegin()
}

type LifecycleRuntimeContextRunGCEnd interface {
	RunGCEnd()
}

type LifecycleRuntimeContextTerminating interface {
	Terminating()
}

type LifecycleRuntimeContextTerminated interface {
	Terminated()
}
