package registry

import (
	"kit.golaxy.org/golaxy/service"
)

// Register 注册服务
func Register(serviceCtx service.Context, service Service, options ...RegisterOption) error {
	return Get(serviceCtx).Register(serviceCtx, service, options...)
}

// Deregister 取消注册服务
func Deregister(serviceCtx service.Context, service Service) error {
	return Get(serviceCtx).Deregister(serviceCtx, service)
}

// GetService 查询服务
func GetService(serviceCtx service.Context, serviceName string) ([]Service, error) {
	return Get(serviceCtx).GetService(serviceCtx, serviceName)
}

// ListServices 查询所有服务
func ListServices(serviceCtx service.Context) ([]Service, error) {
	return Get(serviceCtx).ListServices(serviceCtx)
}

// Watch 获取服务监听器
func Watch(serviceCtx service.Context, serviceName string) (Watcher, error) {
	return Get(serviceCtx).Watch(serviceCtx, serviceName)
}
