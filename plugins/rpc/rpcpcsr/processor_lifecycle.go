package rpcpcsr

import "git.golaxy.org/core/service"

// LifecycleInit RPC处理器生命周期开始
type LifecycleInit interface {
	Init(ctx service.Context)
}

// LifecycleShut RPC处理器生命周期结束
type LifecycleShut interface {
	Shut(ctx service.Context)
}
