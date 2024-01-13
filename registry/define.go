package registry

import (
	"git.golaxy.org/core/define"
)

var (
	plugin = define.DefineServicePluginInterface[IRegistry]()
	Name   = plugin.Name
	Using  = plugin.Using
)
