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
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/framework/addins/gate"
	"git.golaxy.org/framework/addins/rpc/callpath"
	"git.golaxy.org/framework/addins/rpcstack"
)

// ProxyGroup 创建分组代理，用于向分组发送RPC
func ProxyGroup(provider runtime.CurrentContextProvider, name string) GroupProxied {
	if provider == nil {
		exception.Panicf("rpc: %w: provider is nil", core.ErrArgs)
	}
	return GroupProxied{
		svcCtx: service.Current(provider),
		rtCtx:  runtime.Current(provider),
		addr:   gate.CliDetails.DomainMulticast.Join(name),
	}
}

// UntrackedProxyGroup 创建分组代理，不继承RPC调用链，用于向分组发送RPC
func UntrackedProxyGroup(svcCtx service.Context, name string) GroupProxied {
	return GroupProxied{
		svcCtx: svcCtx,
		addr:   gate.CliDetails.DomainMulticast.Join(name),
	}
}

// GroupProxied 分组代理，用于向分组发送RPC
type GroupProxied struct {
	svcCtx service.Context
	rtCtx  runtime.Context
	addr   string
}

// GetName 获取分组名称
func (p GroupProxied) GetName() string {
	name, _ := gate.CliDetails.DomainMulticast.Relative(p.addr)
	return name
}

// GetAddr 获取分组地址
func (p GroupProxied) GetAddr() string {
	return p.addr
}

// CliOnewayRPC 向分组发送单向RPC
func (p GroupProxied) CliOnewayRPC(proc, method string, args ...any) error {
	if p.svcCtx == nil {
		exception.Panic("rpc: svcCtx is nil")
	}

	// 调用链
	cc := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		cc = rpcstack.Using(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Client,
		Script:   proc,
		Method:   method,
	}

	return Using(p.svcCtx).OnewayRPC(p.addr, cc, cp, args...)
}
