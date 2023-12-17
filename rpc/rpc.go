package rpc

import (
	"kit.golaxy.org/golaxy/runtime"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util/option"
	"kit.golaxy.org/golaxy/util/types"
	"kit.golaxy.org/plugins/log"
	"kit.golaxy.org/plugins/util/concurrent"
)

// RPC RPC支持
type RPC interface {
	// RPC RPC调用
	RPC(dst, path string, args ...any) runtime.AsyncRet
	// OneWayRPC 单向RPC调用
	OneWayRPC(dst, path string, args ...any) error
}

func newRPC(settings ...option.Setting[RPCOptions]) RPC {
	return &_RPC{
		options: option.Make(Option{}.Default(), settings...),
	}
}

type _RPC struct {
	options RPCOptions
	ctx     service.Context
}

// InitSP 初始化服务插件
func (r *_RPC) InitSP(ctx service.Context) {
	log.Infof(ctx, "init service plugin %q with %q", plugin.Name, types.AnyFullName(*r))

	r.ctx = ctx
}

// ShutSP 关闭服务插件
func (r *_RPC) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut service plugin %q", plugin.Name)
}

// RPC RPC调用
func (r *_RPC) RPC(dst, path string, args ...any) runtime.AsyncRet {
	for i := range r.options.Deliverers {
		deliverer := r.options.Deliverers[i]

		if !deliverer.Match(r.ctx, dst, path, false) {
			continue
		}

		return deliverer.Request(r.ctx, dst, path, args)
	}

	ret := concurrent.MakeRespAsyncRet()
	ret.Push(concurrent.MakeRet[any](nil, ErrNoDeliverer))

	return ret.Cast()
}

// OneWayRPC 单向RPC调用
func (r *_RPC) OneWayRPC(dst, path string, args ...any) error {
	for i := range r.options.Deliverers {
		deliverer := r.options.Deliverers[i]

		if !deliverer.Match(r.ctx, dst, path, true) {
			continue
		}

		return deliverer.Notify(r.ctx, dst, path, args)
	}

	return ErrNoDeliverer
}
