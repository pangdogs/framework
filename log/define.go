package log

import "kit.golaxy.org/golaxy/define"

var (
	plugin = define.DefinePluginInterface[ILogger]()
	Name   = plugin.Name
	Using  = plugin.Using
)
