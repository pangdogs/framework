package redis

import (
	"kit.golaxy.org/golaxy/define"
	"kit.golaxy.org/plugins/registry"
)

var definePlugin = define.DefineServicePlugin[registry.Registry, RedisOption](NewRedisRegistry)

// Install 安装插件
var Install = definePlugin.Install

// Uninstall 卸载插件
var Uninstall = definePlugin.Uninstall
