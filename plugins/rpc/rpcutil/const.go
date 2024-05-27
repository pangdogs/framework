package rpcutil

import (
	"git.golaxy.org/framework/plugins/gate"
	"git.golaxy.org/framework/plugins/rpc/callpath"
	"git.golaxy.org/framework/plugins/rpcstack"
	"strings"
)

const (
	NoPlugin = ""
	NoComp   = ""
)

func ClientRPCPermValidator(callChain rpcstack.CallChain, cp callpath.CallPath) bool {
	if !gate.CliDetails.InDomain(callChain[0].Src) {
		return true
	}
	return strings.HasPrefix(cp.Method, "C_")
}
