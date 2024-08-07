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
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/plugins/dserv"
	"git.golaxy.org/framework/plugins/rpc"
	"git.golaxy.org/framework/plugins/rpc/callpath"
	"git.golaxy.org/framework/plugins/rpcstack"
)

// ProxyService 代理服务
func ProxyService(ctx service.Context, service ...string) ServiceProxied {
	p := ServiceProxied{
		servCtx: ctx,
	}

	if len(service) > 0 {
		p.service = service[0]
	}

	return p
}

// ServiceProxied 实体服务，用于向服务发送RPC
type ServiceProxied struct {
	servCtx service.Context
	service string
}

// GetService 获取服务名
func (p ServiceProxied) GetService() string {
	return p.service
}

// RPC 向分布式服务指定节点发送RPC
func (p ServiceProxied) RPC(nodeId uid.Id, plugin, method string, args ...any) async.AsyncRet {
	if p.servCtx == nil {
		panic(errors.New("rpc: setting servCtx is nil"))
	}

	// 目标地址
	dst, err := dserv.Using(p.servCtx).MakeNodeAddr(nodeId)
	if err != nil {
		return makeErr(err)
	}

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Service,
		Plugin:   plugin,
		Method:   method,
	}

	return rpc.Using(p.servCtx).RPC(dst, rpcstack.EmptyCallChain, cp.String(), args...)
}

// BalanceRPC 使用负载均衡模式，向分布式服务发送RPC
func (p ServiceProxied) BalanceRPC(plugin, method string, args ...any) async.AsyncRet {
	if p.servCtx == nil {
		panic(errors.New("rpc: setting servCtx is nil"))
	}

	// 目标地址
	var dst string

	if p.service != "" {
		dst = dserv.Using(p.servCtx).MakeBalanceAddr(p.service)
	} else {
		dst = dserv.Using(p.servCtx).GetNodeDetails().GlobalBalanceAddr
	}

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Service,
		Plugin:   plugin,
		Method:   method,
	}

	return rpc.Using(p.servCtx).RPC(dst, rpcstack.EmptyCallChain, cp.String(), args...)
}

// OneWayRPC 向分布式服务指定节点发送单向RPC
func (p ServiceProxied) OneWayRPC(nodeId uid.Id, plugin, method string, args ...any) error {
	if p.servCtx == nil {
		panic(errors.New("rpc: setting servCtx is nil"))
	}

	// 目标地址
	dst, err := dserv.Using(p.servCtx).MakeNodeAddr(nodeId)
	if err != nil {
		return err
	}

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Service,
		Plugin:   plugin,
		Method:   method,
	}

	return rpc.Using(p.servCtx).OneWayRPC(dst, rpcstack.EmptyCallChain, cp.String(), args...)
}

// BalanceOneWayRPC 使用负载均衡模式，向分布式服务发送单向RPC
func (p ServiceProxied) BalanceOneWayRPC(plugin, method string, args ...any) error {
	if p.servCtx == nil {
		panic(errors.New("rpc: setting servCtx is nil"))
	}

	// 目标地址
	var dst string

	if p.service != "" {
		dst = dserv.Using(p.servCtx).MakeBalanceAddr(p.service)
	} else {
		dst = dserv.Using(p.servCtx).GetNodeDetails().GlobalBalanceAddr
	}

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Service,
		Plugin:   plugin,
		Method:   method,
	}

	return rpc.Using(p.servCtx).OneWayRPC(dst, rpcstack.EmptyCallChain, cp.String(), args...)
}

// BroadcastOneWayRPC 使用广播模式，向分布式服务发送单向RPC
func (p ServiceProxied) BroadcastOneWayRPC(excludeSelf bool, plugin, method string, args ...any) error {
	if p.servCtx == nil {
		panic(errors.New("rpc: setting servCtx is nil"))
	}

	// 目标地址
	var dst string

	if p.service != "" {
		dst = dserv.Using(p.servCtx).MakeBroadcastAddr(p.service)
	} else {
		dst = dserv.Using(p.servCtx).GetNodeDetails().GlobalBroadcastAddr
	}

	// 调用路径
	cp := callpath.CallPath{
		Category:   callpath.Service,
		ExcludeSrc: excludeSelf,
		Plugin:     plugin,
		Method:     method,
	}

	return rpc.Using(p.servCtx).OneWayRPC(dst, rpcstack.EmptyCallChain, cp.String(), args...)
}
