package broker

import "kit.golaxy.org/golaxy/define"

var (
	plugin = define.DefineServicePluginInterface[IBroker]()
	Name   = plugin.Name
	Using  = plugin.Using
)
