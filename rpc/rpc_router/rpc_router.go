package rpc_router

import (
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util"
	"kit.golaxy.org/plugins/broker"
	"kit.golaxy.org/plugins/logger"
	"kit.golaxy.org/plugins/registry"
	"kit.golaxy.org/plugins/rpc"
)

func newRPCRouter(options ...RPCRouterOption) rpc.RPCRouter {
	opts := RPCRouterOptions{}
	Option{}.Default()(&opts)

	for i := range options {
		options[i](&opts)
	}

	return &_RPCRouter{
		options: opts,
	}
}

type _RPCRouter struct {
	options  RPCRouterOptions
	ctx      service.Context
	registry registry.Registry
	broker   broker.Broker
	gate     gate.Gate
}

// InitSP 初始化服务插件
func (r *_RPCRouter) InitSP(ctx service.Context) {
	logger.Infof(ctx, "init service plugin %q with %q", definePlugin.Name, util.TypeOfAnyFullName(*r))

	r.ctx = ctx

	r.registry = registry.Fetch(ctx)
	r.broker = broker.Fetch(ctx)
	r.gate = gate.Fetch(ctx)

}

// ShutSP 关闭服务插件
func (r *_RPCRouter) ShutSP(ctx service.Context) {
	logger.Infof(ctx, "shut service plugin %q", definePlugin.Name)
}
