package gtp_gate

import (
	"kit.golaxy.org/golaxy/define"
)

var (
	plugin    = define.DefineServicePlugin[Gate, GateOption](newGate)
	Name      = plugin.Name
	Using     = plugin.Using
	Install   = plugin.Install
	Uninstall = plugin.Uninstall
)
