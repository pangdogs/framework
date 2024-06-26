package rpc

import (
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/plugins/rpc/rpcpcsr"
)

type RPCOptions struct {
	Processors []any
}

var With _Option

type _Option struct{}

func (_Option) Default() option.Setting[RPCOptions] {
	return func(options *RPCOptions) {
		With.Processors(rpcpcsr.NewServiceProcessor(nil))(options)
	}
}

func (_Option) Processors(processors ...any) option.Setting[RPCOptions] {
	return func(options *RPCOptions) {
		options.Processors = processors
	}
}
