package etcd

import (
	"github.com/golaxy-kit/golaxy/define"
	"github.com/golaxy-kit/plugins/registry"
)

var plugin = define.DefineServicePlugin[registry.Registry, EtcdOption](newRegistry)

var Install = plugin.Install

var Uninstall = plugin.Uninstall
