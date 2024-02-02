package dserv

import (
	"git.golaxy.org/core/define"
)

var (
	self      = define.DefineServicePlugin(newDistService)
	Name      = self.Name
	Using     = self.Using
	Install   = self.Install
	Uninstall = self.Uninstall
)
