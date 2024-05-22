package mongodb

import "git.golaxy.org/core/define"

var (
	self      = define.ServicePlugin(newMongoDB)
	Name      = self.Name
	Using     = self.Using
	Install   = self.Install
	Uninstall = self.Uninstall
)
