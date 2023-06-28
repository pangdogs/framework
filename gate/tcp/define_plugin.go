package tcp

import (
	"kit.golaxy.org/golaxy/define"
	"kit.golaxy.org/plugins/gate"
)

var definePlugin = define.DefineServicePlugin[gate.Gate, GateOption](newTcpGate)

// Install 安装插件
var Install = definePlugin.Install

// Uninstall 卸载插件
var Uninstall = definePlugin.Uninstall