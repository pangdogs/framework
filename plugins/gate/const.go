package gate

import "git.golaxy.org/framework/net/netpath"

// CliDetails 客户端地址信息
var CliDetails = netpath.NodeDetails{
	Domain:             "cli",
	NodeSubdomain:      "cli.nd",
	MulticastSubdomain: "cli.mc",
	PathSeparator:      ".",
}
