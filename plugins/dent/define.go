package dent

import "git.golaxy.org/core/define"

var (
	self      = define.RuntimePlugin(newDistEntities)
	Name      = self.Name
	Using     = self.Using
	Install   = self.Install
	Uninstall = self.Uninstall
)
