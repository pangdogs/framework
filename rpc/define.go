package rpc

import (
	"git.golaxy.org/core/define"
)

var (
	plugin    = define.DefineServicePlugin(newRPC)
	Name      = plugin.Name
	Using     = plugin.Using
	Install   = plugin.Install
	Uninstall = plugin.Uninstall
)
