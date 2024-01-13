package distributed

import (
	"git.golaxy.org/core/define"
)

var (
	plugin    = define.DefineServicePlugin(newDistributed)
	Name      = plugin.Name
	Using     = plugin.Using
	Install   = plugin.Install
	Uninstall = plugin.Uninstall
)
