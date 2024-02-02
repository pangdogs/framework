package cache_discovery

import (
	"git.golaxy.org/core/define"
)

var (
	self      = define.DefineServicePlugin(newRegistry)
	Install   = self.Install
	Uninstall = self.Uninstall
)
