package redis_discovery

import (
	"git.golaxy.org/core/define"
)

var (
	plugin    = define.DefineServicePlugin(NewRegistry)
	Install   = plugin.Install
	Uninstall = plugin.Uninstall
)
