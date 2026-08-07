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

// ProxyEntity 使用 provider 所在的服务上下文创建实体 id 的 RPC 代理。
// provider 必须是 service.Context 或实现 runtime.CurrentContextProvider，否则 panic。
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

// EntityProxied 绑定一个实体 ID，用于向承载该实体的服务节点或关联客户端发起 RPC 调用。
type EntityProxied struct {
	svcCtx service.Context
	rtCtx  runtime.Context
	id     uid.Id
}

// RPC 向承载实体的首个指定服务节点发起 RPC；查询失败时返回已携带错误的 Future。
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

// BalanceRPC 从承载实体且服务名匹配的节点中随机选择一个发起 RPC。
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

// GlobalBalanceRPC 从承载实体的全部节点中随机选择一个发起 RPC；excludeSelf 为 true 时排除本节点。
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

// OnewayRPC 向承载实体的首个指定服务节点发起单向 RPC。
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

// BalanceOnewayRPC 从承载实体且服务名匹配的节点中随机选择一个发起单向 RPC。
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

// GlobalBalanceOnewayRPC 从承载实体的全部节点中随机选择一个发起单向 RPC；excludeSelf 为 true 时排除本节点。
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

// BroadcastOnewayRPC 向指定服务中承载该实体的节点广播单向 RPC；excludeSelf 为 true 时排除源节点。
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

// GlobalBroadcastOnewayRPC 向所有服务中承载该实体的节点广播单向 RPC；excludeSelf 为 true 时排除源节点。
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

// CliRPC 向实体 ID 对应的客户端单播地址发起 RPC。
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

// CliOnewayRPC 向实体 ID 对应的客户端单播地址发起单向 RPC。
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
