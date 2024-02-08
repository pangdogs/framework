package conf

import "git.golaxy.org/core/define"

var (
	self      = define.DefinePlugin(newConfig)
	Name      = self.Name
	Using     = self.Using
	Install   = self.Install
	Uninstall = self.Uninstall
)
