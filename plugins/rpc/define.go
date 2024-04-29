package rpc

import (
	"git.golaxy.org/core/define"
)

var (
	self      = define.ServicePlugin(newRPC)
	Name      = self.Name
	Using     = self.Using
	Install   = self.Install
	Uninstall = self.Uninstall
)
