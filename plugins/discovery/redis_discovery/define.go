package redis_discovery

import (
	"git.golaxy.org/core/define"
)

var (
	self      = define.ServicePlugin(NewRegistry)
	Install   = self.Install
	Uninstall = self.Uninstall
)
