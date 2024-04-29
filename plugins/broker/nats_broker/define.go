package nats_broker

import (
	"git.golaxy.org/core/define"
)

var (
	self      = define.ServicePlugin(newBroker)
	Install   = self.Install
	Uninstall = self.Uninstall
)
