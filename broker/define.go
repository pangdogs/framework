package broker

import "kit.golaxy.org/golaxy/define"

var (
	plugin = define.DefineServicePluginInterface[Broker]()
	Name   = plugin.Name
	Using  = plugin.Using
)
