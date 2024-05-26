package framework

type iEC interface {
	GetRuntime() Runtime
	GetService() Service
	IsAlive() bool
}
