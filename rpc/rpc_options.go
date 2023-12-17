package rpc

import "kit.golaxy.org/golaxy/util/option"

type Option struct{}

type RPCOptions struct {
	Deliverers []Deliverer
	Routers    []Router
}

func (Option) Default() option.Setting[RPCOptions] {
	return func(options *RPCOptions) {
		Option{}.Deliverers(&DistributedDeliverer{})
		Option{}.Routers()
	}
}

func (Option) Deliverers(deliverers ...Deliverer) option.Setting[RPCOptions] {
	return func(options *RPCOptions) {
		options.Deliverers = deliverers
	}
}

func (Option) Routers(routers ...Router) option.Setting[RPCOptions] {
	return func(options *RPCOptions) {
		options.Routers = routers
	}
}
