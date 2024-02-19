package rpc

import "git.golaxy.org/core/util/option"

type RPCOptions struct {
	Deliverers  []IDeliverer
	Dispatchers []IDispatcher
}

var With _Option

type _Option struct{}

func (_Option) Default() option.Setting[RPCOptions] {
	return func(options *RPCOptions) {
		With.Deliverers(&DistributedDeliverer{})(options)
		With.Dispatchers(&DistributedDispatcher{})(options)
	}
}

func (_Option) Deliverers(deliverers ...IDeliverer) option.Setting[RPCOptions] {
	return func(options *RPCOptions) {
		options.Deliverers = deliverers
	}
}

func (_Option) Dispatchers(dispatchers ...IDispatcher) option.Setting[RPCOptions] {
	return func(options *RPCOptions) {
		options.Dispatchers = dispatchers
	}
}
