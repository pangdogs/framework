package log

import "git.golaxy.org/core/define"

var (
	plugin = define.DefinePluginInterface[ILogger]()
	Name   = plugin.Name
	Using  = plugin.Using
)
