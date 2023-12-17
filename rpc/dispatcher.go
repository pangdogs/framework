package rpc

import (
	"errors"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/plugins/gap"
)

var (
	ErrUnableToDispatch = errors.New("rpc: unable to dispatch") // 无法分发RPC
)

// Dispatcher RPC分发器
type Dispatcher interface {
	// Match 是否匹配
	Match(ctx service.Context, src string) bool
	// Dispatching 分发消息
	Dispatching(ctx service.Context, src string, msg gap.Msg) error
}
