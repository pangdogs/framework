package dentq

import "git.golaxy.org/core/define"

var (
	self      = define.DefineServicePlugin(newDistEntityQuerier)
	Name      = self.Name
	Using     = self.Using
	Install   = self.Install
	Uninstall = self.Uninstall
)
