package gate

import (
	"git.golaxy.org/core/define"
)

var (
	self      = define.ServicePlugin(newGate)
	Name      = self.Name
	Using     = self.Using
	Install   = self.Install
	Uninstall = self.Uninstall
)
