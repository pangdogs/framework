package rpc

import (
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/plugins/rpc/processor"
	"git.golaxy.org/framework/plugins/rpcstack"
	"git.golaxy.org/framework/util/concurrent"
	"sync/atomic"
)

// IRPC RPC支持
type IRPC interface {
	// RPC RPC调用
	RPC(dst string, callChain rpcstack.CallChain, path string, args ...any) async.AsyncRet
	// OneWayRPC 单向RPC调用
	OneWayRPC(dst string, callChain rpcstack.CallChain, path string, args ...any) error
}

func newRPC(settings ...option.Setting[RPCOptions]) IRPC {
	return &_RPC{
		options: option.Make(With.Default(), settings...),
	}
}

type _RPC struct {
	options    RPCOptions
	servCtx    service.Context
	terminated atomic.Bool
	deliverers []processor.IDeliverer
}

// InitSP 初始化服务插件
func (r *_RPC) InitSP(ctx service.Context) {
	log.Infof(ctx, "init plugin %q", self.Name)

	r.servCtx = ctx

	for _, p := range r.options.Processors {
		if deliverer, ok := p.(processor.IDeliverer); ok {
			r.deliverers = append(r.deliverers, deliverer)
		}
	}

	for _, p := range r.options.Processors {
		if init, ok := p.(processor.LifecycleInit); ok {
			init.Init(r.servCtx)
		}
	}
}

// ShutSP 关闭服务插件
func (r *_RPC) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut plugin %q", self.Name)

	r.terminated.Store(true)

	for _, p := range r.options.Processors {
		if shut, ok := p.(processor.LifecycleShut); ok {
			shut.Shut(r.servCtx)
		}
	}
}

// RPC RPC调用
func (r *_RPC) RPC(dst string, callChain rpcstack.CallChain, path string, args ...any) async.AsyncRet {
	if r.terminated.Load() {
		ret := concurrent.MakeRespAsyncRet()
		ret.Push(async.MakeRet(nil, processor.ErrTerminated))
		return ret.CastAsyncRet()
	}

	if callChain == nil {
		callChain = rpcstack.EmptyCallChain
	}

	for i := range r.deliverers {
		deliverer := r.deliverers[i]

		if !deliverer.Match(r.servCtx, dst, callChain, path, false) {
			continue
		}

		return deliverer.Request(r.servCtx, dst, callChain, path, args)
	}

	ret := concurrent.MakeRespAsyncRet()
	ret.Push(async.MakeRet(nil, processor.ErrUndeliverable))
	return ret.CastAsyncRet()
}

// OneWayRPC 单向RPC调用
func (r *_RPC) OneWayRPC(dst string, callChain rpcstack.CallChain, path string, args ...any) error {
	if r.terminated.Load() {
		return processor.ErrTerminated
	}

	if callChain == nil {
		callChain = rpcstack.EmptyCallChain
	}

	for i := range r.deliverers {
		deliverer := r.deliverers[i]

		if !deliverer.Match(r.servCtx, dst, callChain, path, true) {
			continue
		}

		return deliverer.Notify(r.servCtx, dst, callChain, path, args)
	}

	return processor.ErrUndeliverable
}
