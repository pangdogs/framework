package rpcstack

import "git.golaxy.org/framework/net/gap/variant"

type (
	Call      = variant.Call
	CallChain = variant.CallChain
)

var EmptyCallChain = CallChain{}

type Variables map[string]any

func (m *Variables) Set(k string, v any) {
	if *m == nil {
		*m = map[string]any{}
	}
	(*m)[k] = v
}

func (m Variables) Get(k string) any {
	if m == nil {
		return nil
	}
	return m[k]
}

func (m Variables) Range(fun func(k string, v any) bool) {
	if fun == nil {
		return
	}

	for k, v := range m {
		if !fun(k, v) {
			return
		}
	}
}
