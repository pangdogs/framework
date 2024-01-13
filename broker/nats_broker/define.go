package nats_broker

import (
	"git.golaxy.org/core/define"
)

var (
	plugin    = define.DefineServicePlugin(newBroker)
	Install   = plugin.Install
	Uninstall = plugin.Uninstall
)
