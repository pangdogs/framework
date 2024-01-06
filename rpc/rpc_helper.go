package rpc

import (
	"kit.golaxy.org/golaxy/runtime"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/plugins/gap/variant"
)

// RPC RPC调用
func RPC(servCtx service.Context, dst, path string, args ...any) runtime.AsyncRet {
	return Using(servCtx).RPC(dst, path, args...)
}

func Return[T any](ret runtime.Ret) error {
	*ret.Value.([]any)[0].(*variant.Int)

}

// OneWayRPC 单向RPC调用
func OneWayRPC(servCtx service.Context, dst, path string, args ...any) error {
	return Using(servCtx).OneWayRPC(dst, path, args...)
}
