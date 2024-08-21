/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package rpc

import (
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/plugins/rpc/rpcpcsr"
	"git.golaxy.org/framework/plugins/rpcstack"
	"git.golaxy.org/framework/utils/concurrent"
	"sync/atomic"
)

// IRPC RPC支持
type IRPC interface {
	// RPC RPC调用
	RPC(dst string, callChain rpcstack.CallChain, path string, args ...any) async.AsyncRet
	// OnewayRPC 单向RPC调用
	OnewayRPC(dst string, callChain rpcstack.CallChain, path string, args ...any) error
}

func newRPC(settings ...option.Setting[RPCOptions]) IRPC {
	return &_RPC{
		options: option.Make(With.Default(), settings...),
	}
}

type _RPC struct {
	servCtx    service.Context
	options    RPCOptions
	terminated atomic.Bool
	deliverers []rpcpcsr.IDeliverer
}

// InitSP 初始化服务插件
func (r *_RPC) InitSP(ctx service.Context) {
	log.Infof(ctx, "init plugin %q", self.Name)

	r.servCtx = ctx

	for _, p := range r.options.Processors {
		if deliverer, ok := p.(rpcpcsr.IDeliverer); ok {
			r.deliverers = append(r.deliverers, deliverer)
		}
	}

	for _, p := range r.options.Processors {
		if init, ok := p.(rpcpcsr.LifecycleInit); ok {
			init.Init(r.servCtx)
		}
	}
}

// ShutSP 关闭服务插件
func (r *_RPC) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut plugin %q", self.Name)

	r.terminated.Store(true)

	for _, p := range r.options.Processors {
		if shut, ok := p.(rpcpcsr.LifecycleShut); ok {
			shut.Shut(r.servCtx)
		}
	}
}

// RPC RPC调用
func (r *_RPC) RPC(dst string, callChain rpcstack.CallChain, path string, args ...any) async.AsyncRet {
	if r.terminated.Load() {
		ret := concurrent.MakeRespAsyncRet()
		ret.Push(async.MakeRet(nil, rpcpcsr.ErrTerminated))
		return ret.ToAsyncRet()
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
	ret.Push(async.MakeRet(nil, rpcpcsr.ErrUndeliverable))
	return ret.ToAsyncRet()
}

// OnewayRPC 单向RPC调用
func (r *_RPC) OnewayRPC(dst string, callChain rpcstack.CallChain, path string, args ...any) error {
	if r.terminated.Load() {
		return rpcpcsr.ErrTerminated
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

	return rpcpcsr.ErrUndeliverable
}
