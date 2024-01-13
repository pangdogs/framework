package etcd_dsync

import (
	"git.golaxy.org/core/define"
)

var (
	plugin    = define.DefineServicePlugin(newDSync)
	Install   = plugin.Install
	Uninstall = plugin.Uninstall
)
