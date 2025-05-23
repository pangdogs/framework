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
	"github.com/elliotchance/pie/v2"
)

// ProxyService 创建服务代理，用于向服务发送RPC
func ProxyService(provider runtime.CurrentContextProvider, serviceName ...string) ServiceProxied {
	if provider == nil {
		exception.Panicf("rpc: %w: provider is nil", core.ErrArgs)
	}

	p := ServiceProxied{
		svcCtx:  service.Current(provider),
		rtCtx:   runtime.Current(provider),
		service: pie.First(serviceName),
	}

	return p
}

// UntrackedProxyService 创建服务代理，不继承RPC调用链，用于向服务发送RPC
func UntrackedProxyService(svcCtx service.Context, serviceName ...string) ServiceProxied {
	if svcCtx == nil {
		exception.Panicf("rpc: %w: svcCtx is nil", core.ErrArgs)
	}

	p := ServiceProxied{
		svcCtx:  svcCtx,
		service: pie.First(serviceName),
	}

	return p
}

// ServiceProxied 服务代理，用于向服务发送RPC
type ServiceProxied struct {
	svcCtx  service.Context
	rtCtx   runtime.Context
	service string
}

// GetService 获取服务名
func (p ServiceProxied) GetService() string {
	return p.service
}

// RPC 向分布式服务指定节点发送RPC
func (p ServiceProxied) RPC(nodeId uid.Id, addIn, method string, args ...any) async.AsyncRet {
	if p.svcCtx == nil {
		exception.Panic("rpc: svcCtx is nil")
	}

	// 目标地址
	dst, err := dsvc.Using(p.svcCtx).GetNodeDetails().MakeNodeAddr(nodeId)
	if err != nil {
		return async.Return(async.MakeAsyncRet(), async.MakeRet(nil, err))
	}

	// 调用链
	cc := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		cc = rpcstack.Using(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Service,
		Script:   addIn,
		Method:   method,
	}

	return Using(p.svcCtx).RPC(dst, cc, cp, args...)
}

// BalanceRPC 使用负载均衡模式，向分布式服务发送RPC
func (p ServiceProxied) BalanceRPC(addIn, method string, args ...any) async.AsyncRet {
	if p.svcCtx == nil {
		exception.Panic("rpc: svcCtx is nil")
	}

	// 目标地址
	var dst string

	if p.service != "" {
		dst = dsvc.Using(p.svcCtx).GetNodeDetails().MakeBalanceAddr(p.service)
	} else {
		dst = dsvc.Using(p.svcCtx).GetNodeDetails().GlobalBalanceAddr
	}

	// 调用链
	cc := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		cc = rpcstack.Using(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Service,
		Script:   addIn,
		Method:   method,
	}

	return Using(p.svcCtx).RPC(dst, cc, cp, args...)
}

// OnewayRPC 向分布式服务指定节点发送单向RPC
func (p ServiceProxied) OnewayRPC(nodeId uid.Id, addIn, method string, args ...any) error {
	if p.svcCtx == nil {
		exception.Panic("rpc: svcCtx is nil")
	}

	// 目标地址
	dst, err := dsvc.Using(p.svcCtx).GetNodeDetails().MakeNodeAddr(nodeId)
	if err != nil {
		return err
	}

	// 调用链
	cc := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		cc = rpcstack.Using(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Service,
		Script:   addIn,
		Method:   method,
	}

	return Using(p.svcCtx).OnewayRPC(dst, cc, cp, args...)
}

// BalanceOnewayRPC 使用负载均衡模式，向分布式服务发送单向RPC
func (p ServiceProxied) BalanceOnewayRPC(addIn, method string, args ...any) error {
	if p.svcCtx == nil {
		exception.Panic("rpc: svcCtx is nil")
	}

	// 目标地址
	var dst string

	if p.service != "" {
		dst = dsvc.Using(p.svcCtx).GetNodeDetails().MakeBalanceAddr(p.service)
	} else {
		dst = dsvc.Using(p.svcCtx).GetNodeDetails().GlobalBalanceAddr
	}

	// 调用链
	cc := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		cc = rpcstack.Using(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Service,
		Script:   addIn,
		Method:   method,
	}

	return Using(p.svcCtx).OnewayRPC(dst, cc, cp, args...)
}

// BroadcastOnewayRPC 使用广播模式，向分布式服务发送单向RPC
func (p ServiceProxied) BroadcastOnewayRPC(excludeSelf bool, addIn, method string, args ...any) error {
	if p.svcCtx == nil {
		exception.Panic("rpc: svcCtx is nil")
	}

	// 目标地址
	var dst string

	if p.service != "" {
		dst = dsvc.Using(p.svcCtx).GetNodeDetails().MakeBroadcastAddr(p.service)
	} else {
		dst = dsvc.Using(p.svcCtx).GetNodeDetails().GlobalBroadcastAddr
	}

	// 调用链
	cc := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		cc = rpcstack.Using(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		Category:   callpath.Service,
		ExcludeSrc: excludeSelf,
		Script:     addIn,
		Method:     method,
	}

	return Using(p.svcCtx).OnewayRPC(dst, cc, cp, args...)
}
