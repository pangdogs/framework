package processor

import "git.golaxy.org/core/service"

// LifecycleInit RPC投递器或分发器生命周期开始
type LifecycleInit interface {
	Init(ctx service.Context)
}

// LifecycleShut RPC投递器或分发器生命周期结束
type LifecycleShut interface {
	Shut(ctx service.Context)
}
