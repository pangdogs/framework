package rpc

import (
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
)

// RPC RPC调用
func RPC(servCtx service.Context, dst, path string, args ...any) runtime.AsyncRet {
	return Using(servCtx).RPC(dst, path, args...)
}

// OneWayRPC 单向RPC调用
func OneWayRPC(servCtx service.Context, dst, path string, args ...any) error {
	return Using(servCtx).OneWayRPC(dst, path, args...)
}
