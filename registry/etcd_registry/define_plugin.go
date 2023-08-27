package etcd_registry

import (
	"kit.golaxy.org/golaxy/define"
	"kit.golaxy.org/plugins/registry"
)

var (
	definePlugin = define.DefineServicePlugin[registry.Registry, RegistryOption](newEtcdRegistry)
	// Install 安装插件
	Install = definePlugin.Install
	// Uninstall 卸载插件
	Uninstall = definePlugin.Uninstall
)
