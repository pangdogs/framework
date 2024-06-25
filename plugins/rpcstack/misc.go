package rpcstack

import (
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/net/gap/variant"
)

type (
	Call      = variant.Call
	CallChain = variant.CallChain
)

var EmptyCallChain = CallChain{}

type Variables = generic.UnorderedSliceMap[string, any]
