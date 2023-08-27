package rpc_resolver

import "regexp"

type Option struct{}

type RPCResolverOptions struct {
	EntityPrototype []string // 实体原型
	regexp.Regexp
}

type RPCResolverOption func(options *RPCResolverOptions)

func (Option) Default() RPCResolverOption {
	return func(options *RPCResolverOptions) {

	}
}
