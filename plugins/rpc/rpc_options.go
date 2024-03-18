package rpc

import (
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/framework/plugins/rpc/processor"
)

type RPCOptions struct {
	Deliverers  []processor.IDeliverer
	Dispatchers []processor.IDispatcher
}

var With _Option

type _Option struct{}

func (_Option) Default() option.Setting[RPCOptions] {
	return func(options *RPCOptions) {
		With.Deliverers(processor.NewServiceDeliverer())(options)
		With.Dispatchers(processor.NewServiceDispatcher())(options)
	}
}

func (_Option) Deliverers(deliverers ...processor.IDeliverer) option.Setting[RPCOptions] {
	return func(options *RPCOptions) {
		options.Deliverers = deliverers
	}
}

func (_Option) Dispatchers(dispatchers ...processor.IDispatcher) option.Setting[RPCOptions] {
	return func(options *RPCOptions) {
		options.Dispatchers = dispatchers
	}
}
