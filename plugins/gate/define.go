package gate

import (
	"git.golaxy.org/core/define"
	"git.golaxy.org/framework/net/netpath"
)

var (
	self      = define.DefineServicePlugin(newGate)
	Name      = self.Name
	Using     = self.Using
	Install   = self.Install
	Uninstall = self.Uninstall
)

// CliDetails 客户端地址信息
var CliDetails = netpath.NodeDetails{
	Domain:             "client",
	NodeSubdomain:      "client.node",
	MulticastSubdomain: "client.multicast",
	PathSeparator:      ".",
}
