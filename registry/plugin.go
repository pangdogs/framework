package registry

import (
	"github.com/golaxy-kit/golaxy/define"
)

var plugin = define.DefineServicePluginInterface[Registry]()

var Name = plugin.Name

var Get = plugin.Get

var TryGet = plugin.TryGet
