package zap_log

import (
	"git.golaxy.org/core/define"
)

var (
	plugin    = define.DefinePlugin(newLogger)
	Install   = plugin.Install
	Uninstall = plugin.Uninstall
)
