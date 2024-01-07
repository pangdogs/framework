package rpc

import (
	"kit.golaxy.org/golaxy/runtime"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util/option"
	"kit.golaxy.org/golaxy/util/types"
	"kit.golaxy.org/plugins/log"
	"kit.golaxy.org/plugins/util/concurrent"
)

// IRPC RPC接口
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
	deliverers  []concurrent.Locked[Deliverer]
	dispatchers []concurrent.Locked[Dispatcher]
}

// InitSP 初始化服务插件
func (r *_RPC) InitSP(ctx service.Context) {
	log.Infof(ctx, "init service plugin <%s>:[%s]", plugin.Name, types.AnyFullName(*r))

	r.servCtx = ctx

	for _, d := range r.options.Deliverers {
		r.deliverers = append(r.deliverers, concurrent.MakeLocked(d))
	}

	for _, d := range r.options.Dispatchers {
		r.dispatchers = append(r.dispatchers, concurrent.MakeLocked(d))
	}

	for i := range r.deliverers {
		r.deliverers[i].AutoLock(func(d *Deliverer) {
			init, ok := (*d).(LifecycleInit)
			if ok {
				init.Init(r.servCtx)
			}
		})
	}

	for i := range r.dispatchers {
		r.dispatchers[i].AutoLock(func(d *Dispatcher) {
			init, ok := (*d).(LifecycleInit)
			if ok {
				init.Init(r.servCtx)
			}
		})
	}
}

// ShutSP 关闭服务插件
func (r *_RPC) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut service plugin <%s>:[%s]", plugin.Name, types.AnyFullName(*r))

	for i := range r.deliverers {
		r.deliverers[i].AutoLock(func(d *Deliverer) {
			shut, ok := (*d).(LifecycleShut)
			if ok {
				shut.Shut(r.servCtx)
			}
		})
	}

	for i := range r.dispatchers {
		r.dispatchers[i].AutoLock(func(d *Dispatcher) {
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

		r.deliverers[i].AutoRLock(func(d *Deliverer) {
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

		r.deliverers[i].AutoRLock(func(d *Deliverer) {
			if !(*d).Match(r.servCtx, dst, path, false) {
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
