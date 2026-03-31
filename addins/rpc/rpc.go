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
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/addins/rpc/callpath"
	"git.golaxy.org/framework/addins/rpc/rpcpcsr"
	"git.golaxy.org/framework/addins/rpcstack"
	"go.uber.org/zap"
)

// IRPC RPC支持
type IRPC interface {
	// RPC RPC调用
	RPC(dst string, cc rpcstack.CallChain, cp callpath.CallPath, args ...any) async.Future
	// OnewayRPC 单向RPC调用
	OnewayRPC(dst string, cc rpcstack.CallChain, cp callpath.CallPath, args ...any) error
}

func newRPC(settings ...option.Setting[RPCOptions]) IRPC {
	return &_RPC{
		options: option.New(With.Default(), settings...),
	}
}

type _RPC struct {
	svcCtx     service.Context
	options    RPCOptions
	barrier    generic.Barrier
	deliverers []rpcpcsr.IDeliverer
}

// Init 初始化插件
func (r *_RPC) Init(svcCtx service.Context) {
	log.L(svcCtx).Info("initializing add-in", zap.String("name", AddIn.Name))

	r.svcCtx = svcCtx

	for _, p := range r.options.Processors {
		if deliverer, ok := p.(rpcpcsr.IDeliverer); ok {
			r.deliverers = append(r.deliverers, deliverer)
		}
	}

	for _, p := range r.options.Processors {
		if cb, ok := p.(rpcpcsr.LifecycleInit); ok {
			cb.Init(r.svcCtx)
		}
	}
}

// Shut 关闭插件
func (r *_RPC) Shut(svcCtx service.Context) {
	log.L(svcCtx).Info("shutting down add-in", zap.String("name", AddIn.Name))

	r.barrier.Close()
	r.barrier.Wait()

	for _, p := range r.options.Processors {
		if cb, ok := p.(rpcpcsr.LifecycleShut); ok {
			cb.Shut(r.svcCtx)
		}
	}
}

// RPC RPC调用
func (r *_RPC) RPC(dst string, cc rpcstack.CallChain, cp callpath.CallPath, args ...any) async.Future {
	if !r.barrier.Join(1) {
		return async.Return(async.NewFutureChan(), async.NewResult(nil, rpcpcsr.ErrTerminated))
	}
	defer r.barrier.Done()

	if cc == nil {
		cc = rpcstack.EmptyCallChain
	}

	for i := range r.deliverers {
		deliverer := r.deliverers[i]

		if !deliverer.Match(r.svcCtx, dst, cc, cp, false) {
			continue
		}

		return deliverer.Request(r.svcCtx, dst, cc, cp, args)
	}

	return async.Return(async.NewFutureChan(), async.NewResult(nil, rpcpcsr.ErrUndeliverable))
}

// OnewayRPC 单向RPC调用
func (r *_RPC) OnewayRPC(dst string, cc rpcstack.CallChain, cp callpath.CallPath, args ...any) error {
	if !r.barrier.Join(1) {
		return rpcpcsr.ErrTerminated
	}
	defer r.barrier.Done()

	if cc == nil {
		cc = rpcstack.EmptyCallChain
	}

	for i := range r.deliverers {
		deliverer := r.deliverers[i]

		if !deliverer.Match(r.svcCtx, dst, cc, cp, true) {
			continue
		}

		return deliverer.Notify(r.svcCtx, dst, cc, cp, args)
	}

	return rpcpcsr.ErrUndeliverable
}
