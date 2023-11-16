package redis_registry

import (
	"kit.golaxy.org/golaxy/define"
)

var (
	plugin    = define.DefineServicePlugin(NewRegistry)
	Install   = plugin.Install
	Uninstall = plugin.Uninstall
)
