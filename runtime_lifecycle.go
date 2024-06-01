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
