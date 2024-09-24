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

package rpcutil

import (
	"errors"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/plugins/dentq"
	"git.golaxy.org/framework/plugins/dsvc"
	"git.golaxy.org/framework/plugins/rpc"
	"git.golaxy.org/framework/plugins/rpc/callpath"
	"git.golaxy.org/framework/plugins/rpc/rpcpcsr"
	"git.golaxy.org/framework/plugins/rpcstack"
	"math/rand"
	"slices"
)

// ProxyRuntime 代理运行时
func ProxyRuntime(provider runtime.CurrentContextProvider, entityId uid.Id) RuntimeProxied {
	return RuntimeProxied{
		svcCtx:   service.Current(provider),
		rtCtx:    runtime.Current(provider),
		entityId: entityId,
	}
}

// ConcurrentProxyRuntime 代理运行时
func ConcurrentProxyRuntime(svcCtx service.Context, entityId uid.Id) RuntimeProxied {
	return RuntimeProxied{
		svcCtx:   svcCtx,
		entityId: entityId,
	}
}

// RuntimeProxied 运行时代理，用于向实体的运行时发送RPC
type RuntimeProxied struct {
	svcCtx   service.Context
	rtCtx    runtime.Context
	entityId uid.Id
}

// GetEntityId 获取实体id
func (p RuntimeProxied) GetEntityId() uid.Id {
	return p.entityId
}

// RPC 向分布式实体目标服务的运行时发送RPC
func (p RuntimeProxied) RPC(service, plugin, method string, args ...any) async.AsyncRet {
	if p.svcCtx == nil {
		panic(errors.New("rpc: setting svcCtx is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.svcCtx).GetDistEntity(p.entityId)
	if !ok {
		return makeErr(rpcpcsr.ErrDistEntityNotFound)
	}

	// 查询分布式实体目标服务节点
	nodeIdx := slices.IndexFunc(distEntity.Nodes, func(node dentq.Node) bool {
		return node.Service == service
	})
	if nodeIdx < 0 {
		return makeErr(rpcpcsr.ErrDistEntityNodeNotFound)
	}

	// 调用链
	cc := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		cc = rpcstack.Using(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Runtime,
		Entity:   p.entityId,
		Plugin:   plugin,
		Method:   method,
	}

	return rpc.Using(p.svcCtx).RPC(distEntity.Nodes[nodeIdx].RemoteAddr, cc, cp, args...)
}

// BalanceRPC 使用负载均衡模式，向分布式实体目标服务的运行时发送RPC
func (p RuntimeProxied) BalanceRPC(service, plugin, method string, args ...any) async.AsyncRet {
	if p.svcCtx == nil {
		panic(errors.New("rpc: setting svcCtx is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.svcCtx).GetDistEntity(p.entityId)
	if !ok {
		return makeErr(rpcpcsr.ErrDistEntityNotFound)
	}

	// 统计节点数量
	var count int
	for i := range distEntity.Nodes {
		if distEntity.Nodes[i].Service == service {
			count++
		}
	}
	if count <= 0 {
		return makeErr(rpcpcsr.ErrDistEntityNodeNotFound)
	}

	// 随机目标节点
	var dst string
	offset := rand.Intn(count)

	for i := range distEntity.Nodes {
		if distEntity.Nodes[i].Service == service {
			if offset <= 0 {
				dst = distEntity.Nodes[i].RemoteAddr
				break
			}
			offset--
		}
	}

	// 调用链
	cc := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		cc = rpcstack.Using(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Runtime,
		Entity:   p.entityId,
		Plugin:   plugin,
		Method:   method,
	}

	return rpc.Using(p.svcCtx).RPC(dst, cc, cp, args...)
}

// GlobalBalanceRPC 使用全局负载均衡模式，向分布式实体任意服务的运行时发送RPC
func (p RuntimeProxied) GlobalBalanceRPC(excludeSelf bool, plugin, method string, args ...any) async.AsyncRet {
	if p.svcCtx == nil {
		panic(errors.New("rpc: setting svcCtx is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.svcCtx).GetDistEntity(p.entityId)
	if !ok {
		return makeErr(rpcpcsr.ErrDistEntityNotFound)
	}

	// 随机目标节点
	var dst string

	if excludeSelf {
		if len(distEntity.Nodes) <= 1 {
			return makeErr(rpcpcsr.ErrDistEntityNodeNotFound)
		}

		localAddr := dsvc.Using(p.svcCtx).GetNodeDetails().LocalAddr
		idx := rand.Intn(len(distEntity.Nodes))

		if distEntity.Nodes[idx].RemoteAddr == localAddr {
			idx = (idx + 1) % len(distEntity.Nodes)
		}

		dst = distEntity.Nodes[idx].RemoteAddr

	} else {
		if len(distEntity.Nodes) <= 0 {
			return makeErr(rpcpcsr.ErrDistEntityNodeNotFound)
		}
		dst = distEntity.Nodes[rand.Intn(len(distEntity.Nodes))].RemoteAddr
	}

	// 调用链
	cc := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		cc = rpcstack.Using(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Runtime,
		Entity:   p.entityId,
		Plugin:   plugin,
		Method:   method,
	}

	return rpc.Using(p.svcCtx).RPC(dst, cc, cp, args...)
}

// OnewayRPC 向分布式实体目标服务的运行时发送单向RPC
func (p RuntimeProxied) OnewayRPC(service, plugin, method string, args ...any) error {
	if p.svcCtx == nil {
		panic(errors.New("rpc: setting svcCtx is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.svcCtx).GetDistEntity(p.entityId)
	if !ok {
		return rpcpcsr.ErrDistEntityNotFound
	}

	// 查询分布式实体目标服务节点
	nodeIdx := slices.IndexFunc(distEntity.Nodes, func(node dentq.Node) bool {
		return node.Service == service
	})
	if nodeIdx < 0 {
		return rpcpcsr.ErrDistEntityNodeNotFound
	}

	// 调用链
	cc := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		cc = rpcstack.Using(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Runtime,
		Entity:   p.entityId,
		Plugin:   plugin,
		Method:   method,
	}

	return rpc.Using(p.svcCtx).OnewayRPC(distEntity.Nodes[nodeIdx].RemoteAddr, cc, cp, args...)
}

// BalanceOnewayRPC 使用负载均衡模式，向分布式实体目标服务的运行时发送单向RPC
func (p RuntimeProxied) BalanceOnewayRPC(service, plugin, method string, args ...any) error {
	if p.svcCtx == nil {
		panic(errors.New("rpc: setting svcCtx is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.svcCtx).GetDistEntity(p.entityId)
	if !ok {
		return rpcpcsr.ErrDistEntityNotFound
	}

	// 统计节点数量
	var count int
	for i := range distEntity.Nodes {
		if distEntity.Nodes[i].Service == service {
			count++
		}
	}
	if count <= 0 {
		return rpcpcsr.ErrDistEntityNodeNotFound
	}

	// 随机目标节点
	var dst string
	offset := rand.Intn(count)

	for i := range distEntity.Nodes {
		if distEntity.Nodes[i].Service == service {
			if offset <= 0 {
				dst = distEntity.Nodes[i].RemoteAddr
				break
			}
			offset--
		}
	}

	// 调用链
	cc := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		cc = rpcstack.Using(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Runtime,
		Entity:   p.entityId,
		Plugin:   plugin,
		Method:   method,
	}

	return rpc.Using(p.svcCtx).OnewayRPC(dst, cc, cp, args...)
}

// GlobalBalanceOnewayRPC 使用全局负载均衡模式，向分布式实体任意服务的运行时发送单向RPC
func (p RuntimeProxied) GlobalBalanceOnewayRPC(excludeSelf bool, plugin, method string, args ...any) error {
	if p.svcCtx == nil {
		panic(errors.New("rpc: setting svcCtx is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.svcCtx).GetDistEntity(p.entityId)
	if !ok {
		return rpcpcsr.ErrDistEntityNotFound
	}

	// 随机目标节点
	var dst string

	if excludeSelf {
		if len(distEntity.Nodes) <= 1 {
			return rpcpcsr.ErrDistEntityNodeNotFound
		}

		localAddr := dsvc.Using(p.svcCtx).GetNodeDetails().LocalAddr
		idx := rand.Intn(len(distEntity.Nodes))

		if distEntity.Nodes[idx].RemoteAddr == localAddr {
			idx = (idx + 1) % len(distEntity.Nodes)
		}

		dst = distEntity.Nodes[idx].RemoteAddr

	} else {
		if len(distEntity.Nodes) <= 0 {
			return rpcpcsr.ErrDistEntityNodeNotFound
		}
		dst = distEntity.Nodes[rand.Intn(len(distEntity.Nodes))].RemoteAddr
	}

	// 调用链
	cc := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		cc = rpcstack.Using(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Runtime,
		Entity:   p.entityId,
		Plugin:   plugin,
		Method:   method,
	}

	return rpc.Using(p.svcCtx).OnewayRPC(dst, cc, cp, args...)
}

// BroadcastOnewayRPC 使用广播模式，向分布式实体目标服务的运行时发送单向RPC
func (p RuntimeProxied) BroadcastOnewayRPC(excludeSelf bool, service, plugin, method string, args ...any) error {
	if p.svcCtx == nil {
		panic(errors.New("rpc: setting svcCtx is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.svcCtx).GetDistEntity(p.entityId)
	if !ok {
		return rpcpcsr.ErrDistEntityNotFound
	}

	// 查询分布式实体目标服务节点
	nodeIdx := slices.IndexFunc(distEntity.Nodes, func(node dentq.Node) bool {
		return node.Service == service
	})
	if nodeIdx < 0 {
		return rpcpcsr.ErrDistEntityNodeNotFound
	}

	// 调用链
	cc := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		cc = rpcstack.Using(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		Category:   callpath.Runtime,
		ExcludeSrc: excludeSelf,
		Entity:     p.entityId,
		Plugin:     plugin,
		Method:     method,
	}

	return rpc.Using(p.svcCtx).OnewayRPC(distEntity.Nodes[nodeIdx].BroadcastAddr, cc, cp, args...)
}

// GlobalBroadcastOnewayRPC 使用全局广播模式，向分布式实体所有服务的运行时发送单向RPC
func (p RuntimeProxied) GlobalBroadcastOnewayRPC(excludeSelf bool, plugin, method string, args ...any) error {
	if p.svcCtx == nil {
		panic(errors.New("rpc: setting svcCtx is nil"))
	}

	// 全局广播地址
	dst := dsvc.Using(p.svcCtx).GetNodeDetails().GlobalBroadcastAddr

	// 调用链
	cc := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		cc = rpcstack.Using(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		Category:   callpath.Runtime,
		ExcludeSrc: excludeSelf,
		Entity:     p.entityId,
		Plugin:     plugin,
		Method:     method,
	}

	return rpc.Using(p.svcCtx).OnewayRPC(dst, cc, cp, args...)
}
