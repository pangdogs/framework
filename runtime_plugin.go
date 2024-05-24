package framework

import (
	"git.golaxy.org/core/runtime"
)

type InstallRuntimeLogger interface {
	InstallLogger(ctx runtime.Context)
}

type InstallRuntimeRPCStack interface {
	InstallRPCStack(ctx runtime.Context)
}

type InstallRuntimeDistEntities interface {
	InstallDistEntities(ctx runtime.Context)
}
