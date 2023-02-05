package registry

import (
	"kit.golaxy.org/golaxy/define"
)

var plugin = define.DefineServicePluginInterface[Registry]()

var Name = plugin.Name

var Get = plugin.Get

var TryGet = plugin.TryGet
