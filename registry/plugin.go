package registry

import (
	"kit.golaxy.org/golaxy/define"
)

var plugin = define.DefineServicePluginInterface[Registry]()

// Name 插件名称
var Name = plugin.Name

// Get 获取插件
var Get = plugin.Get

// TryGet 尝试获取插件
var TryGet = plugin.TryGet
