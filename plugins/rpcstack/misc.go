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

type Variables generic.UnorderedSliceMap[string, any]

func (m *Variables) Set(k string, v any) {
	(*generic.UnorderedSliceMap[string, any])(m).Add(k, v)
}

func (m Variables) Get(k string) any {
	v, ok := (generic.UnorderedSliceMap[string, any])(m).Get(k)
	if !ok {
		return nil
	}
	return v
}

func (m Variables) Range(fun generic.Func2[string, any, bool]) {
	for _, kv := range (generic.UnorderedSliceMap[string, any])(m) {
		if !fun.Exec(kv.K, kv.V) {
			return
		}
	}
}

func (m Variables) Each(fun generic.Action2[string, any]) {
	for _, kv := range (generic.UnorderedSliceMap[string, any])(m) {
		fun.Exec(kv.K, kv.V)
	}
}
