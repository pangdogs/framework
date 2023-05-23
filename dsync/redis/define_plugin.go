package redis

import (
	"kit.golaxy.org/golaxy/define"
	"kit.golaxy.org/plugins/dsync"
)

var definePlugin = define.DefineServicePlugin[dsync.DSync, DSyncOption](newRedisDSync)

// Install 安装插件
var Install = definePlugin.Install

// Uninstall 卸载插件
var Uninstall = definePlugin.Uninstall
