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

package rpcstack

import (
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/framework/plugins/log"
)

// IRPCStack RPC调用堆栈支持
type IRPCStack interface {
	iRPCStack
	// CallChain 调用链
	CallChain() CallChain
	// Variables 栈变量
	Variables() *Variables
}

type iRPCStack interface {
	pushCallChain(cc CallChain)
	popCallChain()
}

func newRPCStack(...any) IRPCStack {
	return &_RPCStack{
		callChain: EmptyCallChain,
		variables: nil,
	}
}

type _RPCStack struct {
	rtCtx     runtime.Context
	callChain CallChain
	variables Variables
}

func (r *_RPCStack) InitRP(ctx runtime.Context) {
	log.Debugf(ctx, "init plugin %q", self.Name)
	r.rtCtx = ctx
}

func (r *_RPCStack) ShutRP(ctx runtime.Context) {
	log.Debugf(ctx, "shut plugin %q", self.Name)
}

// CallChain 调用链
func (r *_RPCStack) CallChain() CallChain {
	return r.callChain
}

// Variables 栈变量
func (r *_RPCStack) Variables() *Variables {
	return &r.variables
}

func (r *_RPCStack) pushCallChain(cc CallChain) {
	if cc == nil {
		cc = EmptyCallChain
	}
	r.callChain = cc
	r.variables = nil
}

func (r *_RPCStack) popCallChain() {
	r.callChain = EmptyCallChain
	r.variables = nil
}
