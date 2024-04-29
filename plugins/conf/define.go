package conf

import "git.golaxy.org/core/define"

var (
	self      = define.Plugin(newConfig)
	Name      = self.Name
	Using     = self.Using
	Install   = self.Install
	Uninstall = self.Uninstall
)
