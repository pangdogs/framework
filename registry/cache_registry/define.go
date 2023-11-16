package cache_registry

import (
	"kit.golaxy.org/golaxy/define"
)

var (
	plugin    = define.DefineServicePlugin(newRegistry)
	Install   = plugin.Install
	Uninstall = plugin.Uninstall
)
