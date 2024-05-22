package redisdb

import "git.golaxy.org/core/define"

var (
	self      = define.ServicePlugin(newRedisDB)
	Name      = self.Name
	Using     = self.Using
	Install   = self.Install
	Uninstall = self.Uninstall
)
