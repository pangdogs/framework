package console_logger

import (
	"kit.golaxy.org/golaxy/define"
	"kit.golaxy.org/plugins/log"
)

var (
	plugin    = define.DefinePlugin[log.Logger, LoggerOption](newLogger)
	Install   = plugin.Install
	Uninstall = plugin.Uninstall
)
