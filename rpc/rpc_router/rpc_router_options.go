package rpc_router

type Option struct{}

type RPCRouterOptions struct {
}

type RPCRouterOption func(options *RPCRouterOptions)

func (Option) Default() RPCRouterOption {
	return func(options *RPCRouterOptions) {

	}
}
