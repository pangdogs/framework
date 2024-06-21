package gate

import "git.golaxy.org/framework/net/netpath"

// CliDetails 客户端地址信息
var CliDetails = netpath.NodeDetails{
	Domain:             "cli",
	BroadcastSubdomain: "cli.bc",
	MulticastSubdomain: "cli.mc",
	NodeSubdomain:      "cli.nd",
	PathSeparator:      ".",
}
