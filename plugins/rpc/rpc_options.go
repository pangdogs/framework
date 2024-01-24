package rpc

import "git.golaxy.org/core/util/option"

type Option struct{}

type RPCOptions struct {
	Deliverers  []IDeliverer
	Dispatchers []IDispatcher
}

func (Option) Default() option.Setting[RPCOptions] {
	return func(options *RPCOptions) {
		Option{}.Deliverers(&DistributedDeliverer{})(options)
		Option{}.Dispatchers(&DistributedDispatcher{})(options)
	}
}

func (Option) Deliverers(deliverers ...IDeliverer) option.Setting[RPCOptions] {
	return func(options *RPCOptions) {
		options.Deliverers = deliverers
	}
}

func (Option) Dispatchers(dispatchers ...IDispatcher) option.Setting[RPCOptions] {
	return func(options *RPCOptions) {
		options.Dispatchers = dispatchers
	}
}
