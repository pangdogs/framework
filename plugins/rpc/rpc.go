package rpc

import (
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/plugins/rpc/processor"
	"git.golaxy.org/framework/util/concurrent"
	"sync/atomic"
)

// IRPC RPC支持
type IRPC interface {
	// RPC RPC调用
	RPC(dst, path string, args ...any) runtime.AsyncRet
	// OneWayRPC 单向RPC调用
	OneWayRPC(dst, path string, args ...any) error
}

func newRPC(settings ...option.Setting[RPCOptions]) IRPC {
	return &_RPC{
		options: option.Make(With.Default(), settings...),
	}
}

type _RPC struct {
	options     RPCOptions
	servCtx     service.Context
	terminated  atomic.Bool
	deliverers  []processor.IDeliverer
	dispatchers []processor.IDispatcher
}

// InitSP 初始化服务插件
func (r *_RPC) InitSP(ctx service.Context) {
	log.Infof(ctx, "init plugin %q", self.Name)

	r.servCtx = ctx

	for _, d := range r.options.Deliverers {
		r.deliverers = append(r.deliverers, d)
	}

	for _, d := range r.options.Dispatchers {
		r.dispatchers = append(r.dispatchers, d)
	}

	for _, d := range r.deliverers {
		init, ok := d.(processor.LifecycleInit)
		if ok {
			init.Init(r.servCtx)
		}
	}

	for _, d := range r.dispatchers {
		init, ok := d.(processor.LifecycleInit)
		if ok {
			init.Init(r.servCtx)
		}
	}
}

// ShutSP 关闭服务插件
func (r *_RPC) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut plugin %q", self.Name)

	r.terminated.Store(true)

	for _, d := range r.deliverers {
		shut, ok := d.(processor.LifecycleShut)
		if ok {
			shut.Shut(r.servCtx)
		}
	}

	for _, d := range r.dispatchers {
		shut, ok := d.(processor.LifecycleShut)
		if ok {
			shut.Shut(r.servCtx)
		}
	}
}

// RPC RPC调用
func (r *_RPC) RPC(dst, path string, args ...any) runtime.AsyncRet {
	if r.terminated.Load() {
		ret := concurrent.MakeRespAsyncRet()
		ret.Push(concurrent.MakeRet[any](nil, processor.ErrTerminated))
		return ret.CastAsyncRet()
	}

	for i := range r.deliverers {
		d := r.deliverers[i]

		if !d.Match(r.servCtx, dst, path, false) {
			continue
		}

		return d.Request(r.servCtx, dst, path, args)
	}

	ret := concurrent.MakeRespAsyncRet()
	ret.Push(concurrent.MakeRet[any](nil, processor.ErrNoDeliverer))
	return ret.CastAsyncRet()
}

// OneWayRPC 单向RPC调用
func (r *_RPC) OneWayRPC(dst, path string, args ...any) error {
	if r.terminated.Load() {
		return processor.ErrTerminated
	}

	for i := range r.deliverers {
		d := r.deliverers[i]

		if !d.Match(r.servCtx, dst, path, true) {
			continue
		}

		return d.Notify(r.servCtx, dst, path, args)
	}

	return processor.ErrNoDeliverer
}
