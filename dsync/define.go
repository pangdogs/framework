package dsync

import "kit.golaxy.org/golaxy/define"

var (
	plugin = define.DefineServicePluginInterface[DSync]()
	Name   = plugin.Name
	Using  = plugin.Using
)
