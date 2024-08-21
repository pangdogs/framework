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

package framework

import (
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/framework/plugins/rpc/rpcutil"
)

// RPC 向分布式实体目标服务发送RPC
func (e *EntityBehavior) RPC(service, comp, method string, args ...any) async.AsyncRet {
	return rpcutil.ProxyEntity(e, e.GetId()).RPC(service, comp, method, args...)
}

// BalanceRPC 使用负载均衡模式，向分布式实体目标服务发送RPC
func (e *EntityBehavior) BalanceRPC(service, comp, method string, args ...any) async.AsyncRet {
	return rpcutil.ProxyEntity(e, e.GetId()).BalanceRPC(service, comp, method, args...)
}

// GlobalBalanceRPC 使用全局负载均衡模式，向分布式实体任意服务发送RPC
func (e *EntityBehavior) GlobalBalanceRPC(excludeSelf bool, comp, method string, args ...any) async.AsyncRet {
	return rpcutil.ProxyEntity(e, e.GetId()).GlobalBalanceRPC(excludeSelf, comp, method, args...)
}

// OnewayRPC 向分布式实体目标服务发送单向RPC
func (e *EntityBehavior) OnewayRPC(service, comp, method string, args ...any) error {
	return rpcutil.ProxyEntity(e, e.GetId()).OnewayRPC(service, comp, method, args...)
}

// BalanceOnewayRPC 使用负载均衡模式，向分布式实体目标服务发送单向RPC
func (e *EntityBehavior) BalanceOnewayRPC(service, comp, method string, args ...any) error {
	return rpcutil.ProxyEntity(e, e.GetId()).BalanceOnewayRPC(service, comp, method, args...)
}

// GlobalBalanceOnewayRPC 使用全局负载均衡模式，向分布式实体任意服务发送单向RPC
func (e *EntityBehavior) GlobalBalanceOnewayRPC(excludeSelf bool, comp, method string, args ...any) error {
	return rpcutil.ProxyEntity(e, e.GetId()).GlobalBalanceOnewayRPC(excludeSelf, comp, method, args...)
}

// BroadcastOnewayRPC 使用广播模式，向分布式实体目标服务发送单向RPC
func (e *EntityBehavior) BroadcastOnewayRPC(excludeSelf bool, service, comp, method string, args ...any) error {
	return rpcutil.ProxyEntity(e, e.GetId()).BroadcastOnewayRPC(excludeSelf, service, comp, method, args...)
}

// GlobalBroadcastOnewayRPC 使用全局广播模式，向分布式实体所有服务发送单向RPC
func (e *EntityBehavior) GlobalBroadcastOnewayRPC(excludeSelf bool, comp, method string, args ...any) error {
	return rpcutil.ProxyEntity(e, e.GetId()).GlobalBroadcastOnewayRPC(excludeSelf, comp, method, args...)
}

// CliRPC 向客户端发送RPC
func (e *EntityBehavior) CliRPC(method string, args ...any) async.AsyncRet {
	return rpcutil.ProxyEntity(e, e.GetId()).CliRPC(method, args...)
}

// CliOnewayRPC 向客户端发送单向RPC
func (e *EntityBehavior) CliOnewayRPC(method string, args ...any) error {
	return rpcutil.ProxyEntity(e, e.GetId()).CliOnewayRPC(method, args...)
}

// BroadcastCliOnewayRPC 向包含实体的所有分组发送单向RPC
func (e *EntityBehavior) BroadcastCliOnewayRPC(method string, args ...any) error {
	return rpcutil.ProxyEntity(e, e.GetId()).BroadcastCliOnewayRPC(method, args...)
}
