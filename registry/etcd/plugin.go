package etcd

import (
	"kit.golaxy.org/golaxy/define"
	"kit.golaxy.org/plugins/registry"
)

var plugin = define.DefineServicePlugin[registry.Registry, EtcdOption](newRegistry)

var Install = plugin.Install

var Uninstall = plugin.Uninstall
