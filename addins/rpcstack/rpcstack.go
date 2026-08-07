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
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/net/gap/variant"
	"go.uber.org/zap"
)

type (
	// Call 是调用链中的单个调用节点。
	Call = variant.Call
	// CallChain 是从入口到当前处理器的 RPC 调用节点序列。
	CallChain = variant.CallChain
	// Variables 保存当前 RPC 调用期间的运行时本地变量。
	Variables = generic.UnorderedSliceMap[string, any]
)

// EmptyCallChain 表示当前运行时没有正在处理的 RPC 调用。
var EmptyCallChain = CallChain{}

// IRPCStack 暴露当前运行时正在处理的 RPC 调用链及其临时变量。
// 仅应在所属运行时 goroutine 中访问。
type IRPCStack interface {
	iRPCStack
	// CallChain 返回当前调用链；没有正在处理的 RPC 时返回 EmptyCallChain。
	CallChain() CallChain
	// Variables 返回当前调用的可变变量表；进入下一次调用时会被清空。
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

func (r *_RPCStack) Init(rtCtx runtime.Context) {
	log.L(rtCtx).Info("initializing add-in", zap.String("name", AddIn.Name))
	r.rtCtx = rtCtx
}

func (r *_RPCStack) Shut(rtCtx runtime.Context) {
	log.L(rtCtx).Info("shutting down add-in", zap.String("name", AddIn.Name))
}

// CallChain 返回当前调用链；没有正在处理的 RPC 时返回 EmptyCallChain。
func (r *_RPCStack) CallChain() CallChain {
	return r.callChain
}

// Variables 返回当前调用的可变变量表；进入下一次调用时会被清空。
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
