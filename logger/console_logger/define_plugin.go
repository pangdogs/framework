package console_logger

import (
	"kit.golaxy.org/golaxy/define"
	"kit.golaxy.org/plugins/logger"
)

var (
	definePlugin = define.DefinePlugin[logger.Logger, LoggerOption](newConsoleLogger)
	// Install 安装插件
	Install = definePlugin.Install
	// Uninstall 卸载插件
	Uninstall = definePlugin.Uninstall
)
