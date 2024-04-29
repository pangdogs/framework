package dserv

import (
	"git.golaxy.org/core/define"
)

var (
	self      = define.ServicePlugin(newDistService)
	Name      = self.Name
	Using     = self.Using
	Install   = self.Install
	Uninstall = self.Uninstall
)
