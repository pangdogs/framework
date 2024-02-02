package conf

import "git.golaxy.org/core/define"

var (
	self      = define.DefineServicePlugin(newConfig)
	Name      = self.Name
	Using     = self.Using
	Install   = self.Install
	Uninstall = self.Uninstall
)
