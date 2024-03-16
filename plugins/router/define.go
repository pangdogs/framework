package router

import "git.golaxy.org/core/define"

var (
	self      = define.DefineServicePlugin(newRouter)
	Name      = self.Name
	Using     = self.Using
	Install   = self.Install
	Uninstall = self.Uninstall
)
