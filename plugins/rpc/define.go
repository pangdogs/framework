package rpc

import (
	"git.golaxy.org/core/define"
)

var (
	self      = define.DefineServicePlugin(newRPC)
	Name      = self.Name
	Using     = self.Using
	Install   = self.Install
	Uninstall = self.Uninstall
)
