package rpc

import "git.golaxy.org/core/service"

// LifecycleProcessorInit RPC投递器或分发器生命周期开始
type LifecycleProcessorInit interface {
	Init(ctx service.Context)
}

// LifecycleProcessorShut RPC投递器或分发器生命周期结束
type LifecycleProcessorShut interface {
	Shut(ctx service.Context)
}
