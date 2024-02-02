package etcd_discovery

import (
	"git.golaxy.org/core/define"
)

var (
	self      = define.DefineServicePlugin(NewRegistry)
	Install   = self.Install
	Uninstall = self.Uninstall
)
