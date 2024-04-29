package router

import "git.golaxy.org/core/define"

var (
	self      = define.ServicePlugin(newRouter)
	Name      = self.Name
	Using     = self.Using
	Install   = self.Install
	Uninstall = self.Uninstall
)
