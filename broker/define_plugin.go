package broker

import "kit.golaxy.org/golaxy/define"

var definePlugin = define.DefineServicePluginInterface[Broker]()

// Name 插件名称
var Name = definePlugin.Name

// Get 获取插件
var Get = definePlugin.Get

// TryGet 尝试获取插件
var TryGet = definePlugin.TryGet
