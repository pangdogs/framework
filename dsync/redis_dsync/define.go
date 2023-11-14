package redis_dsync

import (
	"kit.golaxy.org/golaxy/define"
)

var (
	plugin    = define.DefineServicePlugin(newDSync)
	Install   = plugin.Install
	Uninstall = plugin.Uninstall
)
