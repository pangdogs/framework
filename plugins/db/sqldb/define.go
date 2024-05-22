package sqldb

import "git.golaxy.org/core/define"

var (
	self      = define.ServicePlugin(newSQLDB)
	Name      = self.Name
	Using     = self.Using
	Install   = self.Install
	Uninstall = self.Uninstall
)
