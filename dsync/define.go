package dsync

import "kit.golaxy.org/golaxy/define"

var (
	plugin = define.DefineServicePluginInterface[IDistSync]()
	Name   = plugin.Name
	Using  = plugin.Using
)
