package app

import "git.golaxy.org/core/service"

type InstallLogger interface {
	InstallLogger(ctx service.Context)
}

type InstallConfig interface {
	InstallConfig(ctx service.Context)
}

type InstallBroker interface {
	InstallBroker(ctx service.Context)
}

type InstallRegistry interface {
	InstallRegistry(ctx service.Context)
}

type InstallDistSync interface {
	InstallDistSync(ctx service.Context)
}

type InstallDistService interface {
	InstallDistService(ctx service.Context)
}

type InstallRPC interface {
	InstallRPC(ctx service.Context)
}

type InstallDistEntityQuerier interface {
	InstallDistEntityQuerier(ctx service.Context)
}
