package rpc_resolver

import (
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/plugins/rpc"
	"net"
	"sync"
)

func newRPCResolver() rpc.RPCResolver {
	opts := GateOptions{}
	Option{}.Default()(&opts)

	for i := range options {
		options[i](&opts)
	}

	return &_GtpGate{
		options: opts,
	}
}

type _RPCResolver struct {
	options      GateOptions
	ctx          service.Context
	listeners    []net.Listener
	sessionMap   sync.Map
	sessionCount int64
}
