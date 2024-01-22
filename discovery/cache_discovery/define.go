package cache_discovery

import (
	"git.golaxy.org/core/define"
)

var (
	plugin    = define.DefineServicePlugin(newRegistry)
	Install   = plugin.Install
	Uninstall = plugin.Uninstall
)
