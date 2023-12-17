package rpc

import (
	"errors"
	"kit.golaxy.org/golaxy/runtime"
	"kit.golaxy.org/golaxy/service"
)

var (
	ErrNoDeliverer = errors.New("rpc: no deliverer") // 没有匹配的投递器
)

// Deliverer RPC投递器，用于将RPC投递至目标
type Deliverer interface {
	// Match 是否匹配
	Match(ctx service.Context, dst, path string, oneWay bool) bool
	// Request 请求
	Request(ctx service.Context, dst, path string, args []any) runtime.AsyncRet
	// Notify 通知
	Notify(ctx service.Context, dst, path string, args []any) error
}
