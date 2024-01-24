package broker

import "git.golaxy.org/core/define"

var (
	plugin = define.DefineServicePluginInterface[IBroker]()
	Name   = plugin.Name
	Using  = plugin.Using
)
