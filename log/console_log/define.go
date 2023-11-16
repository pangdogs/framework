package console_log

import (
	"kit.golaxy.org/golaxy/define"
)

var (
	plugin    = define.DefinePlugin(newLogger)
	Install   = plugin.Install
	Uninstall = plugin.Uninstall
)
