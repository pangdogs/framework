package dent

import "git.golaxy.org/core/define"

var (
	plugin    = define.DefineRuntimePlugin(newDistEntities)
	Name      = plugin.Name
	Using     = plugin.Using
	Install   = plugin.Install
	Uninstall = plugin.Uninstall
)
