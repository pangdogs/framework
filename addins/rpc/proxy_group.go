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

// ProxyGroup 使用 provider 所在的服务上下文创建客户端组 name 的 RPC 代理。
// provider 必须是 service.Context 或实现 runtime.CurrentContextProvider，否则 panic。
func ProxyGroup(provider any, name string) GroupProxied {
	if provider == nil {
		exception.Panicf("rpc: %w: provider is nil", core.ErrArgs)
	}
	p := GroupProxied{
		addr: gate.ClientDetails.DomainMulticast.Join(name),
	}
	switch x := provider.(type) {
	case runtime.CurrentContextProvider:
		p.svcCtx = service.Current(x)
		p.rtCtx = runtime.Current(x)
	case service.Context:
		p.svcCtx = x
	default:
		exception.Panicf("rpc: %w: invalid provider type", core.ErrArgs)
	}
	return p
}

// GroupProxied 绑定一个客户端组播地址。
type GroupProxied struct {
	svcCtx service.Context
	rtCtx  runtime.Context
	addr   string
}

// CliOnewayRPC 向组内客户端广播单向 RPC。
func (p GroupProxied) CliOnewayRPC(script, method string, args ...any) error {
	if p.svcCtx == nil {
		exception.Panic("rpc: svcCtx is nil")
	}

	// 调用链
	cc := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		cc = rpcstack.AddIn.Require(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		TargetKind: callpath.Client,
		Script:     script,
		Method:     method,
	}

	return AddIn.Require(p.svcCtx).OnewayRPC(p.addr, cc, cp, args...)
}
