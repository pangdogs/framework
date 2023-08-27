package rpc_resolver

import (
	"kit.golaxy.org/golaxy/define"
	"kit.golaxy.org/plugins/rpc"
)

var (
	definePlugin = define.DefineServicePlugin[rpc.RPCResolver]()
	// Install 安装插件
	Install = definePlugin.Install
	// Uninstall 卸载插件
	Uninstall = definePlugin.Uninstall
)
