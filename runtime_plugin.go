package framework

type InstallRuntimeLogger interface {
	InstallLogger(inst IRuntimeInstance)
}

type InstallRuntimeRPCStack interface {
	InstallRPCStack(inst IRuntimeInstance)
}

type InstallRuntimeDistEntityRegistry interface {
	InstallDistEntityRegistry(inst IRuntimeInstance)
}
