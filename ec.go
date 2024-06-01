package framework

type iEC interface {
	GetRuntime() IRuntimeInstance
	GetService() IServiceInstance
	IsAlive() bool
}
