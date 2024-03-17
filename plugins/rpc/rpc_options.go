package rpc

import (
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/framework/plugins/rpc/processors"
)

type RPCOptions struct {
	Deliverers  []IProcessorDeliverer
	Dispatchers []IProcessorDispatcher
}

var With _Option

type _Option struct{}

func (_Option) Default() option.Setting[RPCOptions] {
	return func(options *RPCOptions) {
		With.Deliverers(processors.NewServiceDeliverer())(options)
		With.Dispatchers(processors.NewServiceDispatcher())(options)
	}
}

func (_Option) Deliverers(deliverers ...IProcessorDeliverer) option.Setting[RPCOptions] {
	return func(options *RPCOptions) {
		options.Deliverers = deliverers
	}
}

func (_Option) Dispatchers(dispatchers ...IProcessorDispatcher) option.Setting[RPCOptions] {
	return func(options *RPCOptions) {
		options.Dispatchers = dispatchers
	}
}
