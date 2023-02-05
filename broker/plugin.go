package broker

import "kit.golaxy.org/golaxy/define"

var plugin = define.DefineServicePluginInterface[Broker]()

var Name = plugin.Name

var Get = plugin.Get

var TryGet = plugin.TryGet
