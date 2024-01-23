package dserv

import (
	"git.golaxy.org/core/define"
)

var (
	plugin    = define.DefineServicePlugin(newDistService)
	Name      = plugin.Name
	Using     = plugin.Using
	Install   = plugin.Install
	Uninstall = plugin.Uninstall
)
