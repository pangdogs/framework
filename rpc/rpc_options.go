package rpc

import "kit.golaxy.org/golaxy/util/option"

type Option struct{}

type RPCOptions struct {
	Deliverers  []Deliverer
	Dispatchers []Dispatcher
}

func (Option) Default() option.Setting[RPCOptions] {
	return func(options *RPCOptions) {
		Option{}.Deliverers(&DistributedDeliverer{})
		Option{}.Dispatchers(&DistributedDispatcher{})
	}
}

func (Option) Deliverers(deliverers ...Deliverer) option.Setting[RPCOptions] {
	return func(options *RPCOptions) {
		options.Deliverers = deliverers
	}
}

func (Option) Dispatchers(dispatchers ...Dispatcher) option.Setting[RPCOptions] {
	return func(options *RPCOptions) {
		options.Dispatchers = dispatchers
	}
}
