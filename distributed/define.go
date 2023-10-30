package distributed

import (
	"kit.golaxy.org/golaxy/define"
)

var (
	plugin    = define.DefineServicePlugin[Distributed, DistributedOption](newDistributed)
	Name      = plugin.Name
	Using     = plugin.Using
	Install   = plugin.Install
	Uninstall = plugin.Uninstall
)
