package rpc

import (
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/util/concurrent"
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
		options: option.Make(Option{}.Default(), settings...),
	}
}

type _RPC struct {
	options     RPCOptions
	servCtx     service.Context
	deliverers  []concurrent.RWLocked[IDeliverer]
	dispatchers []concurrent.RWLocked[IDispatcher]
}

// InitSP 初始化服务插件
func (r *_RPC) InitSP(ctx service.Context) {
	log.Infof(ctx, "init plugin %q", self.Name)

	r.servCtx = ctx

	for _, d := range r.options.Deliverers {
		r.deliverers = append(r.deliverers, concurrent.MakeRWLocked(d))
	}

	for _, d := range r.options.Dispatchers {
		r.dispatchers = append(r.dispatchers, concurrent.MakeRWLocked(d))
	}

	for i := range r.deliverers {
		r.deliverers[i].AutoLock(func(d *IDeliverer) {
			init, ok := (*d).(LifecycleInit)
			if ok {
				init.Init(r.servCtx)
			}
		})
	}

	for i := range r.dispatchers {
		r.dispatchers[i].AutoLock(func(d *IDispatcher) {
			init, ok := (*d).(LifecycleInit)
			if ok {
				init.Init(r.servCtx)
			}
		})
	}
}

// ShutSP 关闭服务插件
func (r *_RPC) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut plugin %q", self.Name)

	for i := range r.deliverers {
		r.deliverers[i].AutoLock(func(d *IDeliverer) {
			shut, ok := (*d).(LifecycleShut)
			if ok {
				shut.Shut(r.servCtx)
			}
		})
	}

	for i := range r.dispatchers {
		r.dispatchers[i].AutoLock(func(d *IDispatcher) {
			shut, ok := (*d).(LifecycleShut)
			if ok {
				shut.Shut(r.servCtx)
			}
		})
	}
}

// RPC RPC调用
func (r *_RPC) RPC(dst, path string, args ...any) runtime.AsyncRet {
	for i := range r.deliverers {
		var ret runtime.AsyncRet

		r.deliverers[i].AutoRLock(func(d *IDeliverer) {
			if !(*d).Match(r.servCtx, dst, path, false) {
				return
			}
			ret = (*d).Request(r.servCtx, dst, path, args)
		})

		if ret != nil {
			return ret
		}
	}

	ret := concurrent.MakeRespAsyncRet()
	ret.Push(concurrent.MakeRet[any](nil, ErrNoDeliverer))

	return ret.CastAsyncRet()
}

// OneWayRPC 单向RPC调用
func (r *_RPC) OneWayRPC(dst, path string, args ...any) error {
	for i := range r.deliverers {
		var b bool
		var err error

		r.deliverers[i].AutoRLock(func(d *IDeliverer) {
			if !(*d).Match(r.servCtx, dst, path, true) {
				return
			}

			b = true
			err = (*d).Notify(r.servCtx, dst, path, args)
		})

		if b {
			return err
		}
	}

	return ErrNoDeliverer
}
