package dsync

import "git.golaxy.org/core/define"

var (
	plugin = define.DefineServicePluginInterface[IDistSync]()
	Name   = plugin.Name
	Using  = plugin.Using
)
