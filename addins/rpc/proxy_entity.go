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
	"math/rand"
	"slices"

	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/addins/dent"
	"git.golaxy.org/framework/addins/dsvc"
	"git.golaxy.org/framework/addins/gate"
	"git.golaxy.org/framework/addins/rpc/callpath"
	"git.golaxy.org/framework/addins/rpc/rpcpcsr"
	"git.golaxy.org/framework/addins/rpcstack"
)

// ProxyEntity 创建实体代理，用于向实体发送RPC
func ProxyEntity(provider any, id uid.Id) EntityProxied {
	if provider == nil {
		exception.Panicf("rpc: %w: provider is nil", core.ErrArgs)
	}
	p := EntityProxied{
		id: id,
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

// EntityProxied 实体代理，用于向实体发送RPC
type EntityProxied struct {
	svcCtx service.Context
	rtCtx  runtime.Context
	id     uid.Id
}

// RPC 向分布式实体目标服务发送RPC
func (p EntityProxied) RPC(service, comp, method string, args ...any) async.Future {
	if p.svcCtx == nil {
		exception.Panic("rpc: svcCtx is nil")
	}

	// 查询分布式实体信息
	distEntity, ok := dent.QuerierAddIn.Require(p.svcCtx).GetDistEntity(p.id)
	if !ok {
		return async.Return(async.NewFutureChan(), async.NewResult(nil, rpcpcsr.ErrDistEntityNotFound))
	}

	// 查询分布式实体目标服务节点
	nodeIdx := slices.IndexFunc(distEntity.Nodes, func(node dent.Node) bool {
		return node.Service == service
	})
	if nodeIdx < 0 {
		return async.Return(async.NewFutureChan(), async.NewResult(nil, rpcpcsr.ErrDistEntityNodeNotFound))
	}

	// 调用链
	cc := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		cc = rpcstack.AddIn.Require(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		TargetKind: callpath.Entity,
		Id:         p.id,
		Script:     comp,
		Method:     method,
	}

	return AddIn.Require(p.svcCtx).RPC(distEntity.Nodes[nodeIdx].RemoteAddr, cc, cp, args...)
}

// BalanceRPC 使用负载均衡模式，向分布式实体目标服务发送RPC
func (p EntityProxied) BalanceRPC(service, comp, method string, args ...any) async.Future {
	if p.svcCtx == nil {
		exception.Panic("rpc: svcCtx is nil")
	}

	// 查询分布式实体信息
	distEntity, ok := dent.QuerierAddIn.Require(p.svcCtx).GetDistEntity(p.id)
	if !ok {
		return async.Return(async.NewFutureChan(), async.NewResult(nil, rpcpcsr.ErrDistEntityNotFound))
	}

	// 统计节点数量
	var count int
	for i := range distEntity.Nodes {
		if distEntity.Nodes[i].Service == service {
			count++
		}
	}
	if count <= 0 {
		return async.Return(async.NewFutureChan(), async.NewResult(nil, rpcpcsr.ErrDistEntityNodeNotFound))
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
		cc = rpcstack.AddIn.Require(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		TargetKind: callpath.Entity,
		Id:         p.id,
		Script:     comp,
		Method:     method,
	}

	return AddIn.Require(p.svcCtx).RPC(dst, cc, cp, args...)
}

// GlobalBalanceRPC 使用全局负载均衡模式，向分布式实体任意服务发送RPC
func (p EntityProxied) GlobalBalanceRPC(excludeSelf bool, comp, method string, args ...any) async.Future {
	if p.svcCtx == nil {
		exception.Panic("rpc: svcCtx is nil")
	}

	// 查询分布式实体信息
	distEntity, ok := dent.QuerierAddIn.Require(p.svcCtx).GetDistEntity(p.id)
	if !ok {
		return async.Return(async.NewFutureChan(), async.NewResult(nil, rpcpcsr.ErrDistEntityNotFound))
	}

	// 随机目标节点
	var dst string

	if excludeSelf {
		if len(distEntity.Nodes) <= 1 {
			return async.Return(async.NewFutureChan(), async.NewResult(nil, rpcpcsr.ErrDistEntityNodeNotFound))
		}

		localAddr := dsvc.AddIn.Require(p.svcCtx).NodeDetails().LocalAddr
		idx := rand.Intn(len(distEntity.Nodes))

		if distEntity.Nodes[idx].RemoteAddr == localAddr {
			idx = (idx + 1) % len(distEntity.Nodes)
		}

		dst = distEntity.Nodes[idx].RemoteAddr

	} else {
		if len(distEntity.Nodes) <= 0 {
			return async.Return(async.NewFutureChan(), async.NewResult(nil, rpcpcsr.ErrDistEntityNodeNotFound))
		}
		dst = distEntity.Nodes[rand.Intn(len(distEntity.Nodes))].RemoteAddr
	}

	// 调用链
	cc := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		cc = rpcstack.AddIn.Require(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		TargetKind: callpath.Entity,
		Id:         p.id,
		Script:     comp,
		Method:     method,
	}

	return AddIn.Require(p.svcCtx).RPC(dst, cc, cp, args...)
}

// OnewayRPC 向分布式实体目标服务发送单向RPC
func (p EntityProxied) OnewayRPC(service, comp, method string, args ...any) error {
	if p.svcCtx == nil {
		exception.Panic("rpc: svcCtx is nil")
	}

	// 查询分布式实体信息
	distEntity, ok := dent.QuerierAddIn.Require(p.svcCtx).GetDistEntity(p.id)
	if !ok {
		return rpcpcsr.ErrDistEntityNotFound
	}

	// 查询分布式实体目标服务节点
	nodeIdx := slices.IndexFunc(distEntity.Nodes, func(node dent.Node) bool {
		return node.Service == service
	})
	if nodeIdx < 0 {
		return rpcpcsr.ErrDistEntityNodeNotFound
	}

	// 调用链
	cc := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		cc = rpcstack.AddIn.Require(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		TargetKind: callpath.Entity,
		Id:         p.id,
		Script:     comp,
		Method:     method,
	}

	return AddIn.Require(p.svcCtx).OnewayRPC(distEntity.Nodes[nodeIdx].RemoteAddr, cc, cp, args...)
}

// BalanceOnewayRPC 使用负载均衡模式，向分布式实体目标服务发送单向RPC
func (p EntityProxied) BalanceOnewayRPC(service, comp, method string, args ...any) error {
	if p.svcCtx == nil {
		exception.Panic("rpc: svcCtx is nil")
	}

	// 查询分布式实体信息
	distEntity, ok := dent.QuerierAddIn.Require(p.svcCtx).GetDistEntity(p.id)
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
		cc = rpcstack.AddIn.Require(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		TargetKind: callpath.Entity,
		Id:         p.id,
		Script:     comp,
		Method:     method,
	}

	return AddIn.Require(p.svcCtx).OnewayRPC(dst, cc, cp, args...)
}

// GlobalBalanceOnewayRPC 使用全局负载均衡模式，向分布式实体任意服务发送单向RPC
func (p EntityProxied) GlobalBalanceOnewayRPC(excludeSelf bool, comp, method string, args ...any) error {
	if p.svcCtx == nil {
		exception.Panic("rpc: svcCtx is nil")
	}

	// 查询分布式实体信息
	distEntity, ok := dent.QuerierAddIn.Require(p.svcCtx).GetDistEntity(p.id)
	if !ok {
		return rpcpcsr.ErrDistEntityNotFound
	}

	// 随机目标节点
	var dst string

	if excludeSelf {
		if len(distEntity.Nodes) <= 1 {
			return rpcpcsr.ErrDistEntityNodeNotFound
		}

		localAddr := dsvc.AddIn.Require(p.svcCtx).NodeDetails().LocalAddr
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
		cc = rpcstack.AddIn.Require(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		TargetKind: callpath.Entity,
		Id:         p.id,
		Script:     comp,
		Method:     method,
	}

	return AddIn.Require(p.svcCtx).OnewayRPC(dst, cc, cp, args...)
}

// BroadcastOnewayRPC 使用广播模式，向分布式实体目标服务发送单向RPC
func (p EntityProxied) BroadcastOnewayRPC(excludeSelf bool, service, comp, method string, args ...any) error {
	if p.svcCtx == nil {
		exception.Panic("rpc: svcCtx is nil")
	}

	// 查询分布式实体信息
	distEntity, ok := dent.QuerierAddIn.Require(p.svcCtx).GetDistEntity(p.id)
	if !ok {
		return rpcpcsr.ErrDistEntityNotFound
	}

	// 查询分布式实体目标服务节点
	nodeIdx := slices.IndexFunc(distEntity.Nodes, func(node dent.Node) bool {
		return node.Service == service
	})
	if nodeIdx < 0 {
		return rpcpcsr.ErrDistEntityNodeNotFound
	}

	// 调用链
	cc := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		cc = rpcstack.AddIn.Require(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		TargetKind: callpath.Entity,
		ExcludeSrc: excludeSelf,
		Id:         p.id,
		Script:     comp,
		Method:     method,
	}

	return AddIn.Require(p.svcCtx).OnewayRPC(distEntity.Nodes[nodeIdx].BroadcastAddr, cc, cp, args...)
}

// GlobalBroadcastOnewayRPC 使用全局广播模式，向分布式实体所有服务发送单向RPC
func (p EntityProxied) GlobalBroadcastOnewayRPC(excludeSelf bool, comp, method string, args ...any) error {
	if p.svcCtx == nil {
		exception.Panic("rpc: svcCtx is nil")
	}

	// 全局广播地址
	dst := dsvc.AddIn.Require(p.svcCtx).NodeDetails().GlobalBroadcastAddr

	// 调用链
	cc := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		cc = rpcstack.AddIn.Require(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		TargetKind: callpath.Entity,
		ExcludeSrc: excludeSelf,
		Id:         p.id,
		Script:     comp,
		Method:     method,
	}

	return AddIn.Require(p.svcCtx).OnewayRPC(dst, cc, cp, args...)
}

// CliRPC 向客户端发送RPC
func (p EntityProxied) CliRPC(script, method string, args ...any) async.Future {
	if p.svcCtx == nil {
		exception.Panic("rpc: svcCtx is nil")
	}

	// 客户端地址
	dst := gate.ClientDetails.DomainUnicast.Join(p.id.String())

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

	return AddIn.Require(p.svcCtx).RPC(dst, cc, cp, args...)
}

// CliOnewayRPC 向客户端发送单向RPC
func (p EntityProxied) CliOnewayRPC(script, method string, args ...any) error {
	if p.svcCtx == nil {
		exception.Panic("rpc: svcCtx is nil")
	}

	// 客户端地址
	dst := gate.ClientDetails.DomainUnicast.Join(p.id.String())

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

	return AddIn.Require(p.svcCtx).OnewayRPC(dst, cc, cp, args...)
}
