package rpcutil

import (
	"git.golaxy.org/framework/plugins/gate"
	"git.golaxy.org/framework/plugins/rpc/callpath"
	"git.golaxy.org/framework/plugins/rpcstack"
	"strings"
)

const (
	NoPlugin = "" // 不使用插件
	NoComp   = "" // 不使用组件
)

// CliRPCPermValidator 默认的客户端RPC请求权限验证器，强制客户端只能RPC调用前缀为C_的函数
func CliRPCPermValidator(callChain rpcstack.CallChain, cp callpath.CallPath) bool {
	if !gate.CliDetails.InDomain(callChain[0].Src) {
		return true
	}
	return strings.HasPrefix(cp.Method, "C_")
}
