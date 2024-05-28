package dentr

import "git.golaxy.org/core/define"

var (
	self      = define.RuntimePlugin(newDistEntityRegistry)
	Name      = self.Name
	Using     = self.Using
	Install   = self.Install
	Uninstall = self.Uninstall
)
