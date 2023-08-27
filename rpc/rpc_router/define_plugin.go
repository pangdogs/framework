package rpc_router

import (
	"kit.golaxy.org/golaxy/define"
	"kit.golaxy.org/plugins/rpc"
)

var (
	definePlugin = define.DefineServicePlugin[rpc.RPCRouter, RPCRouterOption](newRPCRouter)
	// Install 安装插件
	Install = definePlugin.Install
	// Uninstall 卸载插件
	Uninstall = definePlugin.Uninstall
)
