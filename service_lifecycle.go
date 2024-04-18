package framework

import "git.golaxy.org/core/service"

type LifecycleServiceBuilt interface {
	Built(ctx service.Context)
}

type LifecycleServiceBirth interface {
	Birth(ctx service.Context)
}

type LifecycleServiceStarting interface {
	Starting(ctx service.Context)
}

type LifecycleServiceStarted interface {
	Started(ctx service.Context)
}

type LifecycleServiceTerminating interface {
	Terminating(ctx service.Context)
}

type LifecycleServiceTerminated interface {
	Terminated(ctx service.Context)
}
