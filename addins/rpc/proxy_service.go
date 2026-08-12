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
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/addins/dsvc"
	"git.golaxy.org/framework/addins/rpc/callpath"
	"git.golaxy.org/framework/addins/rpcstack"
)

// ProxyService 使用 provider 所在的服务上下文创建服务 RPC 代理。
// provider 必须是 service.Context 或实现 runtime.CurrentContextProvider，否则 panic。
func ProxyService(provider any) ServiceProxied {
	if provider == nil {
		exception.Panicf("rpc: %w: provider is nil", core.ErrArgs)
	}
	p := ServiceProxied{}
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

// ServiceProxied 用于调用分布式服务节点中的服务插件方法。
type ServiceProxied struct {
	svcCtx service.Context
	rtCtx  runtime.Context
}

// RPC 向 nodeId 标识的服务节点发起 RPC；地址构造失败时返回已携带错误的 Future。
func (p ServiceProxied) RPC(nodeId uid.Id, addIn, method string, args ...any) async.Future {
	if p.svcCtx == nil {
		exception.Panic("rpc: svcCtx is nil")
	}

	// 目标地址
	dst, err := dsvc.AddIn.Require(p.svcCtx).NodeDetails().MakeNodeAddr(nodeId)
	if err != nil {
		return async.Rejected(err)
	}

	// 调用链
	cc := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		cc = rpcstack.AddIn.Require(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		TargetKind: callpath.Service,
		Script:     addIn,
		Method:     method,
	}

	return AddIn.Require(p.svcCtx).RPC(dst, cc, cp, args...)
}

// BalanceRPC 向指定服务名的负载均衡地址发起 RPC；service 为空时使用全局负载均衡地址。
func (p ServiceProxied) BalanceRPC(service, addIn, method string, args ...any) async.Future {
	if p.svcCtx == nil {
		exception.Panic("rpc: svcCtx is nil")
	}

	// 目标地址
	var dst string

	if service != "" {
		dst = dsvc.AddIn.Require(p.svcCtx).NodeDetails().MakeBalanceAddr(service)
	} else {
		dst = dsvc.AddIn.Require(p.svcCtx).NodeDetails().GlobalBalanceAddr
	}

	// 调用链
	cc := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		cc = rpcstack.AddIn.Require(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		TargetKind: callpath.Service,
		Script:     addIn,
		Method:     method,
	}

	return AddIn.Require(p.svcCtx).RPC(dst, cc, cp, args...)
}

// OnewayRPC 向 nodeId 标识的服务节点发起单向 RPC。
func (p ServiceProxied) OnewayRPC(nodeId uid.Id, addIn, method string, args ...any) error {
	if p.svcCtx == nil {
		exception.Panic("rpc: svcCtx is nil")
	}

	// 目标地址
	dst, err := dsvc.AddIn.Require(p.svcCtx).NodeDetails().MakeNodeAddr(nodeId)
	if err != nil {
		return err
	}

	// 调用链
	cc := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		cc = rpcstack.AddIn.Require(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		TargetKind: callpath.Service,
		Script:     addIn,
		Method:     method,
	}

	return AddIn.Require(p.svcCtx).OnewayRPC(dst, cc, cp, args...)
}

// BalanceOnewayRPC 向指定服务名的负载均衡地址发起单向 RPC；service 为空时使用全局负载均衡地址。
func (p ServiceProxied) BalanceOnewayRPC(service, addIn, method string, args ...any) error {
	if p.svcCtx == nil {
		exception.Panic("rpc: svcCtx is nil")
	}

	// 目标地址
	var dst string

	if service != "" {
		dst = dsvc.AddIn.Require(p.svcCtx).NodeDetails().MakeBalanceAddr(service)
	} else {
		dst = dsvc.AddIn.Require(p.svcCtx).NodeDetails().GlobalBalanceAddr
	}

	// 调用链
	cc := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		cc = rpcstack.AddIn.Require(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		TargetKind: callpath.Service,
		Script:     addIn,
		Method:     method,
	}

	return AddIn.Require(p.svcCtx).OnewayRPC(dst, cc, cp, args...)
}

// BroadcastOnewayRPC 向指定服务名广播单向 RPC；service 为空时全局广播，excludeSelf 为 true 时排除源节点。
func (p ServiceProxied) BroadcastOnewayRPC(excludeSelf bool, service, addIn, method string, args ...any) error {
	if p.svcCtx == nil {
		exception.Panic("rpc: svcCtx is nil")
	}

	// 目标地址
	var dst string

	if service != "" {
		dst = dsvc.AddIn.Require(p.svcCtx).NodeDetails().MakeBroadcastAddr(service)
	} else {
		dst = dsvc.AddIn.Require(p.svcCtx).NodeDetails().GlobalBroadcastAddr
	}

	// 调用链
	cc := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		cc = rpcstack.AddIn.Require(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		TargetKind: callpath.Service,
		ExcludeSrc: excludeSelf,
		Script:     addIn,
		Method:     method,
	}

	return AddIn.Require(p.svcCtx).OnewayRPC(dst, cc, cp, args...)
}
