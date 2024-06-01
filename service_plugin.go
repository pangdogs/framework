package framework

type InstallServiceLogger interface {
	InstallLogger(inst IServiceInstance)
}

type InstallServiceConfig interface {
	InstallConfig(inst IServiceInstance)
}

type InstallServiceBroker interface {
	InstallBroker(inst IServiceInstance)
}

type InstallServiceRegistry interface {
	InstallRegistry(inst IServiceInstance)
}

type InstallServiceDistSync interface {
	InstallDistSync(inst IServiceInstance)
}

type InstallServiceDistService interface {
	InstallDistService(inst IServiceInstance)
}

type InstallServiceRPC interface {
	InstallRPC(inst IServiceInstance)
}

type InstallServiceDistEntityQuerier interface {
	InstallDistEntityQuerier(inst IServiceInstance)
}
