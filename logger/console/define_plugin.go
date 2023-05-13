package console

import (
	"kit.golaxy.org/golaxy/define"
	"kit.golaxy.org/plugins/logger"
)

var definePlugin = define.DefinePlugin[logger.Logger, Option](newConsoleLogger)

// Install 安装插件
var Install = definePlugin.Install

// Uninstall 卸载插件
var Uninstall = definePlugin.Uninstall
