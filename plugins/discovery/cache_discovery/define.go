package cache_discovery

import (
	"git.golaxy.org/core/define"
)

var (
	self      = define.ServicePlugin(newRegistry)
	Install   = self.Install
	Uninstall = self.Uninstall
)
