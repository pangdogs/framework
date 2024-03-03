package framework

import (
	"git.golaxy.org/core/service"
)

type SetupServiceContextComposite interface {
	MakeContextComposite() service.Context
}

type LifecycleServiceContextBirth interface {
	Birth()
}

type LifecycleServiceContextInit interface {
	Init()
}

type LifecycleServiceContextStarting interface {
	Starting()
}

type LifecycleServiceContextStarted interface {
	Started()
}

type LifecycleServiceContextTerminating interface {
	Terminating()
}

type LifecycleServiceContextTerminated interface {
	Terminated()
}
