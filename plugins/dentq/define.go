package dentq

import "git.golaxy.org/core/define"

var (
	plugin    = define.DefineServicePlugin(newDistEntityQuerier)
	Name      = plugin.Name
	Using     = plugin.Using
	Install   = plugin.Install
	Uninstall = plugin.Uninstall
)
