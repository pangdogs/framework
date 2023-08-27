package rpc

import "kit.golaxy.org/golaxy/define"

var (
	defineRPCRouterPlugin   = define.DefineServicePluginInterface[RPCRouter]()
	defineRPCResolverPlugin = define.DefineServicePluginInterface[RPCResolver]()
)
