package console

import (
	"kit.golaxy.org/golaxy/define"
	"kit.golaxy.org/plugins/logger"
)

var plugin = define.DefineServicePlugin[logger.Logger, ConsoleOption](newConsoleLogger)

// Install 安装插件
var Install = plugin.Install

// Uninstall 卸载插件
var Uninstall = plugin.Uninstall
