package nats_broker

import (
	"kit.golaxy.org/golaxy/define"
	"kit.golaxy.org/plugins/broker"
)

var (
	plugin    = define.DefineServicePlugin[broker.Broker, BrokerOption](newBroker)
	Install   = plugin.Install
	Uninstall = plugin.Uninstall
)
