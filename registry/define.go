package registry

import (
	"kit.golaxy.org/golaxy/define"
)

var (
	plugin = define.DefineServicePluginInterface[Registry]()
	Name   = plugin.Name
	Using  = plugin.Using
)
