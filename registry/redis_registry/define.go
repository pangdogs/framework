package redis_registry

import (
	"kit.golaxy.org/golaxy/define"
	"kit.golaxy.org/plugins/registry"
)

var (
	plugin    = define.DefineServicePlugin[registry.Registry, RegistryOption](NewRegistry)
	Install   = plugin.Install
	Uninstall = plugin.Uninstall
)
