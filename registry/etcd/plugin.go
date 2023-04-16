package etcd

import (
	"kit.golaxy.org/golaxy/define"
	"kit.golaxy.org/plugins/registry"
)

var plugin = define.DefineServicePlugin[registry.Registry, EtcdOption](newEtcdRegistry)

// Install 安装插件
var Install = plugin.Install

// Uninstall 卸载插件
var Uninstall = plugin.Uninstall
