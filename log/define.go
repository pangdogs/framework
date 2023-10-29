package log

import "kit.golaxy.org/golaxy/define"

var (
	plugin = define.DefinePluginInterface[Logger]()
	Name   = plugin.Name
	Using  = plugin.Using
)
