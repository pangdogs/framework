package rpcstack

import (
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/framework/plugins/log"
)

// IRPCStack RPC调用堆栈支持
type IRPCStack interface {
	iRPCStack
	// CallChain 调用链
	CallChain() CallChain
	// Variables 栈变量
	Variables() *Variables
}

type iRPCStack interface {
	pushCallChain(callChain CallChain)
	popCallChain()
}

func newRPCStack(...any) IRPCStack {
	return &_RPCStack{
		callChain: EmptyCallChain,
		variables: nil,
	}
}

type _RPCStack struct {
	rtCtx     runtime.Context
	callChain CallChain
	variables Variables
}

func (r *_RPCStack) InitRP(ctx runtime.Context) {
	log.Debugf(ctx, "init plugin %q", self.Name)
	r.rtCtx = ctx
}

func (r *_RPCStack) ShutRP(ctx runtime.Context) {
	log.Debugf(ctx, "shut plugin %q", self.Name)
}

// CallChain 调用链
func (r *_RPCStack) CallChain() CallChain {
	return r.callChain
}

// Variables 栈变量
func (r *_RPCStack) Variables() *Variables {
	return &r.variables
}

func (r *_RPCStack) pushCallChain(callChain CallChain) {
	if callChain == nil {
		callChain = EmptyCallChain
	}
	r.callChain = callChain
	r.variables = nil
}

func (r *_RPCStack) popCallChain() {
	r.callChain = EmptyCallChain
	r.variables = nil
}
