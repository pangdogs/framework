package logger

import "kit.golaxy.org/golaxy/define"

var plugin = define.DefineServicePluginInterface[Logger]()

var Name = plugin.Name

var Get = plugin.Get

var TryGet = plugin.TryGet
