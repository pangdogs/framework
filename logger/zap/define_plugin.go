package zap

import (
	"kit.golaxy.org/golaxy/define"
	"kit.golaxy.org/plugins/logger"
)

var definePlugin = define.DefinePlugin[logger.Logger, ZapOption](newZapLogger)

// Install 安装插件
var Install = definePlugin.Install

// Uninstall 卸载插件
var Uninstall = definePlugin.Uninstall
