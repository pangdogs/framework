package nats_broker

import (
	"kit.golaxy.org/golaxy/define"
	"kit.golaxy.org/plugins/broker"
)

var (
	definePlugin = define.DefineServicePlugin[broker.Broker, BrokerOption](newNatsBroker)
	// Install 安装插件
	Install = definePlugin.Install
	// Uninstall 卸载插件
	Uninstall = definePlugin.Uninstall
)
