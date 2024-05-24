package rpcstack

// Deprecated: UnsafeFrame 访问RPC调用堆栈支持内部方法
func UnsafeRPCStack(r IRPCStack) _UnsafeRPCStack {
	return _UnsafeRPCStack{
		IRPCStack: r,
	}
}

type _UnsafeRPCStack struct {
	IRPCStack
}

func (ur _UnsafeRPCStack) PushCallChain(callChain CallChain) {
	ur.pushCallChain(callChain)
}

func (ur _UnsafeRPCStack) PopCallChain() {
	ur.popCallChain()
}
