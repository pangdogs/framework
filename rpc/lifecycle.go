package rpc

import "kit.golaxy.org/golaxy/service"

// LifecycleInit 生命周期开始
type LifecycleInit interface {
	Init(ctx service.Context)
}

// LifecycleShut 生命周期结束
type LifecycleShut interface {
	Shut(ctx service.Context)
}
