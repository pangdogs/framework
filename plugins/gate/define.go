package gate

import (
	"git.golaxy.org/core/define"
)

var (
	self      = define.DefineServicePlugin(newGate)
	Name      = self.Name
	Using     = self.Using
	Install   = self.Install
	Uninstall = self.Uninstall
)

const (
	ClientDomain             = "client"           // 客户端主域
	ClientNodeSubdomain      = "client.node"      // 客户端节点地址子域
	ClientMulticastSubdomain = "client.multicast" // 客户端组播地址子域
	ClientPathSeparator      = "."                // 客户端地址路径分隔符
)
