package gtp_gate

import (
	"git.golaxy.org/core/define"
)

var (
	plugin    = define.DefineServicePlugin(newGate)
	Name      = plugin.Name
	Using     = plugin.Using
	Install   = plugin.Install
	Uninstall = plugin.Uninstall
)
