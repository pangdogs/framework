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
	"git.golaxy.org/framework/plugins/dserv"
	"git.golaxy.org/framework/plugins/gate"
	"git.golaxy.org/framework/plugins/rpc"
	"git.golaxy.org/framework/plugins/rpc/callpath"
	"git.golaxy.org/framework/plugins/rpc/rpcpcsr"
	"git.golaxy.org/framework/plugins/rpcstack"
	"math/rand"
	"slices"
)

func makeErr(err error) async.AsyncRet {
	asyncRet := make(chan async.Ret, 1)
	asyncRet <- async.MakeRet(nil, err)
	close(asyncRet)
	return asyncRet
}

// ProxyEntity 代理实体
func ProxyEntity(ctx runtime.CurrentContextProvider, id uid.Id) EntityProxied {
	return EntityProxied{
		servCtx: service.Current(ctx),
		rtCtx:   runtime.Current(ctx),
		id:      id,
	}
}

// ConcurrentProxyEntity 代理实体
func ConcurrentProxyEntity(ctx service.Context, id uid.Id) EntityProxied {
	return EntityProxied{
		servCtx: ctx,
		id:      id,
	}
}

// EntityProxied 实体代理，用于向实体发送RPC
type EntityProxied struct {
	servCtx service.Context
	rtCtx   runtime.Context
	id      uid.Id
}

// GetId 获取实体id
func (p EntityProxied) GetId() uid.Id {
	return p.id
}

// RPC 向分布式实体目标服务发送RPC
func (p EntityProxied) RPC(service, comp, method string, args ...any) async.AsyncRet {
	if p.servCtx == nil {
		panic(errors.New("rpc: setting servCtx is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.servCtx).GetDistEntity(p.id)
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
	callChain := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		callChain = rpcstack.Using(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		Category:  callpath.Entity,
		EntityId:  p.id,
		Component: comp,
		Method:    method,
	}

	return rpc.Using(p.servCtx).RPC(distEntity.Nodes[nodeIdx].RemoteAddr, callChain, cp.String(), args...)
}

// BalanceRPC 使用负载均衡模式，向分布式实体目标服务发送RPC
func (p EntityProxied) BalanceRPC(service, comp, method string, args ...any) async.AsyncRet {
	if p.servCtx == nil {
		panic(errors.New("rpc: setting servCtx is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.servCtx).GetDistEntity(p.id)
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
	callChain := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		callChain = rpcstack.Using(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		Category:  callpath.Entity,
		EntityId:  p.id,
		Component: comp,
		Method:    method,
	}

	return rpc.Using(p.servCtx).RPC(dst, callChain, cp.String(), args...)
}

// GlobalBalanceRPC 使用全局负载均衡模式，向分布式实体任意服务发送RPC
func (p EntityProxied) GlobalBalanceRPC(excludeSelf bool, comp, method string, args ...any) async.AsyncRet {
	if p.servCtx == nil {
		panic(errors.New("rpc: setting servCtx is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.servCtx).GetDistEntity(p.id)
	if !ok {
		return makeErr(rpcpcsr.ErrDistEntityNotFound)
	}

	// 随机目标节点
	var dst string

	if excludeSelf {
		if len(distEntity.Nodes) <= 1 {
			return makeErr(rpcpcsr.ErrDistEntityNodeNotFound)
		}

		localAddr := dserv.Using(p.servCtx).GetNodeDetails().LocalAddr
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
	callChain := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		callChain = rpcstack.Using(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		Category:  callpath.Entity,
		EntityId:  p.id,
		Component: comp,
		Method:    method,
	}

	return rpc.Using(p.servCtx).RPC(dst, callChain, cp.String(), args...)
}

// OneWayRPC 向分布式实体目标服务发送单向RPC
func (p EntityProxied) OneWayRPC(service, comp, method string, args ...any) error {
	if p.servCtx == nil {
		panic(errors.New("rpc: setting servCtx is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.servCtx).GetDistEntity(p.id)
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
	callChain := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		callChain = rpcstack.Using(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		Category:  callpath.Entity,
		EntityId:  p.id,
		Component: comp,
		Method:    method,
	}

	return rpc.Using(p.servCtx).OneWayRPC(distEntity.Nodes[nodeIdx].RemoteAddr, callChain, cp.String(), args...)
}

// BalanceOneWayRPC 使用负载均衡模式，向分布式实体目标服务发送单向RPC
func (p EntityProxied) BalanceOneWayRPC(service, comp, method string, args ...any) error {
	if p.servCtx == nil {
		panic(errors.New("rpc: setting servCtx is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.servCtx).GetDistEntity(p.id)
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
	callChain := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		callChain = rpcstack.Using(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		Category:  callpath.Entity,
		EntityId:  p.id,
		Component: comp,
		Method:    method,
	}

	return rpc.Using(p.servCtx).OneWayRPC(dst, callChain, cp.String(), args...)
}

// GlobalBalanceOneWayRPC 使用全局负载均衡模式，向分布式实体任意服务发送单向RPC
func (p EntityProxied) GlobalBalanceOneWayRPC(excludeSelf bool, comp, method string, args ...any) error {
	if p.servCtx == nil {
		panic(errors.New("rpc: setting servCtx is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.servCtx).GetDistEntity(p.id)
	if !ok {
		return rpcpcsr.ErrDistEntityNotFound
	}

	// 随机目标节点
	var dst string

	if excludeSelf {
		if len(distEntity.Nodes) <= 1 {
			return rpcpcsr.ErrDistEntityNodeNotFound
		}

		localAddr := dserv.Using(p.servCtx).GetNodeDetails().LocalAddr
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
	callChain := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		callChain = rpcstack.Using(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		Category:  callpath.Entity,
		EntityId:  p.id,
		Component: comp,
		Method:    method,
	}

	return rpc.Using(p.servCtx).OneWayRPC(dst, callChain, cp.String(), args...)
}

// BroadcastOneWayRPC 使用广播模式，向分布式实体目标服务发送单向RPC
func (p EntityProxied) BroadcastOneWayRPC(excludeSelf bool, service, comp, method string, args ...any) error {
	if p.servCtx == nil {
		panic(errors.New("rpc: setting servCtx is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.servCtx).GetDistEntity(p.id)
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
	callChain := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		callChain = rpcstack.Using(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		Category:   callpath.Entity,
		ExcludeSrc: excludeSelf,
		EntityId:   p.id,
		Component:  comp,
		Method:     method,
	}

	return rpc.Using(p.servCtx).OneWayRPC(distEntity.Nodes[nodeIdx].BroadcastAddr, callChain, cp.String(), args...)
}

// GlobalBroadcastOneWayRPC 使用全局广播模式，向分布式实体所有服务发送单向RPC
func (p EntityProxied) GlobalBroadcastOneWayRPC(excludeSelf bool, comp, method string, args ...any) error {
	if p.servCtx == nil {
		panic(errors.New("rpc: setting servCtx is nil"))
	}

	// 全局广播地址
	dst := dserv.Using(p.servCtx).GetNodeDetails().GlobalBroadcastAddr

	// 调用链
	callChain := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		callChain = rpcstack.Using(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		Category:   callpath.Entity,
		ExcludeSrc: excludeSelf,
		EntityId:   p.id,
		Component:  comp,
		Method:     method,
	}

	return rpc.Using(p.servCtx).OneWayRPC(dst, callChain, cp.String(), args...)
}

// CliRPC 向客户端发送RPC
func (p EntityProxied) CliRPC(method string, args ...any) async.AsyncRet {
	return p.CliRPCToEntity(uid.Nil, method, args...)
}

// CliRPCToEntity 向客户端实体发送RPC
func (p EntityProxied) CliRPCToEntity(entityId uid.Id, method string, args ...any) async.AsyncRet {
	if p.servCtx == nil {
		panic(errors.New("rpc: setting servCtx is nil"))
	}

	// 客户端地址
	dst := gate.CliDetails.DomainUnicast.Join(p.id.String())

	// 调用链
	callChain := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		callChain = rpcstack.Using(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Client,
		EntityId: entityId,
		Method:   method,
	}

	return rpc.Using(p.servCtx).RPC(dst, callChain, cp.String(), args...)
}

// OneWayCliRPC 向客户端发送单向RPC
func (p EntityProxied) OneWayCliRPC(method string, args ...any) error {
	return p.OneWayCliRPCToEntity(uid.Nil, method, args...)
}

// OneWayCliRPCToEntity 向客户端实体发送单向RPC
func (p EntityProxied) OneWayCliRPCToEntity(entityId uid.Id, method string, args ...any) error {
	if p.servCtx == nil {
		panic(errors.New("rpc: setting servCtx is nil"))
	}

	// 客户端地址
	dst := gate.CliDetails.DomainUnicast.Join(p.id.String())

	// 调用链
	callChain := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		callChain = rpcstack.Using(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Client,
		EntityId: entityId,
		Method:   method,
	}

	return rpc.Using(p.servCtx).OneWayRPC(dst, callChain, cp.String(), args...)
}

// GroupOneWayCliRPC 向包含实体的分组发送单向RPC
func (p EntityProxied) GroupOneWayCliRPC(method string, args ...any) error {
	if p.servCtx == nil {
		panic(errors.New("rpc: setting servCtx is nil"))
	}

	// 客户端地址
	dst := gate.CliDetails.DomainBroadcast.Join(p.id.String())

	// 调用链
	callChain := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		callChain = rpcstack.Using(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Client,
		EntityId: p.id,
		Method:   method,
	}

	return rpc.Using(p.servCtx).OneWayRPC(dst, callChain, cp.String(), args...)
}
