package cache_registry

import (
	"kit.golaxy.org/golaxy/define"
	"kit.golaxy.org/plugins/registry"
)

var (
	plugin    = define.DefineServicePlugin[registry.Registry, RegistryOption](newRegistry)
	Install   = plugin.Install
	Uninstall = plugin.Uninstall
)
