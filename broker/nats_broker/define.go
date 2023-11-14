package nats_broker

import (
	"kit.golaxy.org/golaxy/define"
)

var (
	plugin    = define.DefineServicePlugin(newBroker)
	Install   = plugin.Install
	Uninstall = plugin.Uninstall
)
