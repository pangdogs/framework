package framework

import "git.golaxy.org/core/service"

type InstallServiceLogger interface {
	InstallLogger(ctx service.Context)
}

type InstallServiceConfig interface {
	InstallConfig(ctx service.Context)
}

type InstallServiceBroker interface {
	InstallBroker(ctx service.Context)
}

type InstallServiceRegistry interface {
	InstallRegistry(ctx service.Context)
}

type InstallServiceDistSync interface {
	InstallDistSync(ctx service.Context)
}

type InstallServiceDistService interface {
	InstallDistService(ctx service.Context)
}

type InstallServiceRPC interface {
	InstallRPC(ctx service.Context)
}

type InstallServiceDistEntityQuerier interface {
	InstallDistEntityQuerier(ctx service.Context)
}
