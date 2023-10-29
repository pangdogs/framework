package etcd_dsync

import (
	"kit.golaxy.org/golaxy/define"
	"kit.golaxy.org/plugins/dsync"
)

var (
	plugin    = define.DefineServicePlugin[dsync.DSync, DSyncOption](newDSync)
	Install   = plugin.Install
	Uninstall = plugin.Uninstall
)
