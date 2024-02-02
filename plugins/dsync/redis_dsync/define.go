package redis_dsync

import (
	"git.golaxy.org/core/define"
)

var (
	self      = define.DefineServicePlugin(newDSync)
	Install   = self.Install
	Uninstall = self.Uninstall
)
