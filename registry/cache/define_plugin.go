package cache

import (
	"kit.golaxy.org/golaxy/define"
	"kit.golaxy.org/plugins/registry"
)

var definePlugin = define.DefineServicePlugin[registry.Registry, Option](newCacheRegistry)

// Install 安装插件
var Install = definePlugin.Install

// Uninstall 卸载插件
var Uninstall = definePlugin.Uninstall
