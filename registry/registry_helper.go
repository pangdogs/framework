package registry

import (
	"context"
	"kit.golaxy.org/golaxy/service"
	"time"
)

// Register 注册服务
func Register(servCtx service.Context, ctx context.Context, service Service, ttl time.Duration) error {
	return Using(servCtx).Register(ctx, service, ttl)
}

// Deregister 取消注册服务
func Deregister(servCtx service.Context, ctx context.Context, service Service) error {
	return Using(servCtx).Deregister(ctx, service)
}

// GetService 查询服务
func GetService(servCtx service.Context, ctx context.Context, serviceName string) ([]Service, error) {
	return Using(servCtx).GetService(ctx, serviceName)
}

// ListServices 查询所有服务
func ListServices(servCtx service.Context, ctx context.Context) ([]Service, error) {
	return Using(servCtx).ListServices(ctx)
}

// Watch 获取服务监听器
func Watch(servCtx service.Context, ctx context.Context, serviceName string) (Watcher, error) {
	return Using(servCtx).Watch(ctx, serviceName)
}
