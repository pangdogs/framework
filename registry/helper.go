package registry

import (
	"context"
	"github.com/galaxy-kit/galaxy-go/service"
)

func Register(serviceCtx service.Context, ctx context.Context, service Service, options ...WithRegisterOption) error {
	return Plugin.Get(serviceCtx).Register(ctx, service, options...)
}

func Deregister(serviceCtx service.Context, ctx context.Context, service Service) error {
	return Plugin.Get(serviceCtx).Deregister(ctx, service)
}

func GetService(serviceCtx service.Context, ctx context.Context, serviceName string) ([]Service, error) {
	return Plugin.Get(serviceCtx).GetService(ctx, serviceName)
}

func ListServices(serviceCtx service.Context, ctx context.Context) ([]Service, error) {
	return Plugin.Get(serviceCtx).ListServices(ctx)
}

func Watch(serviceCtx service.Context, ctx context.Context, serviceName string) (Watcher, error) {
	return Plugin.Get(serviceCtx).Watch(ctx, serviceName)
}
