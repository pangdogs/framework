package framework

type LifecycleServiceBuilt interface {
	Built(inst IServiceInstance)
}

type LifecycleServiceBirth interface {
	Birth(inst IServiceInstance)
}

type LifecycleServiceStarting interface {
	Starting(inst IServiceInstance)
}

type LifecycleServiceStarted interface {
	Started(inst IServiceInstance)
}

type LifecycleServiceTerminating interface {
	Terminating(inst IServiceInstance)
}

type LifecycleServiceTerminated interface {
	Terminated(inst IServiceInstance)
}
