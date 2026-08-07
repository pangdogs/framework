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
	"git.golaxy.org/framework/addins/rpc"
)

// RPC 向承载当前实体的首个指定服务节点发起 RPC。
func (e *EntityBehavior) RPC(service, comp, method string, args ...any) async.Future {
	return rpc.ProxyEntity(e, e.Id()).RPC(service, comp, method, args...)
}

// BalanceRPC 从承载当前实体且服务名匹配的节点中随机选择一个发起 RPC。
func (e *EntityBehavior) BalanceRPC(service, comp, method string, args ...any) async.Future {
	return rpc.ProxyEntity(e, e.Id()).BalanceRPC(service, comp, method, args...)
}

// GlobalBalanceRPC 从承载当前实体的全部节点中随机选择一个发起 RPC；excludeSelf 为 true 时排除本节点。
func (e *EntityBehavior) GlobalBalanceRPC(excludeSelf bool, comp, method string, args ...any) async.Future {
	return rpc.ProxyEntity(e, e.Id()).GlobalBalanceRPC(excludeSelf, comp, method, args...)
}

// OnewayRPC 向承载当前实体的首个指定服务节点发起单向 RPC。
func (e *EntityBehavior) OnewayRPC(service, comp, method string, args ...any) error {
	return rpc.ProxyEntity(e, e.Id()).OnewayRPC(service, comp, method, args...)
}

// BalanceOnewayRPC 从承载当前实体且服务名匹配的节点中随机选择一个发起单向 RPC。
func (e *EntityBehavior) BalanceOnewayRPC(service, comp, method string, args ...any) error {
	return rpc.ProxyEntity(e, e.Id()).BalanceOnewayRPC(service, comp, method, args...)
}

// GlobalBalanceOnewayRPC 从承载当前实体的全部节点中随机选择一个发起单向 RPC；excludeSelf 为 true 时排除本节点。
func (e *EntityBehavior) GlobalBalanceOnewayRPC(excludeSelf bool, comp, method string, args ...any) error {
	return rpc.ProxyEntity(e, e.Id()).GlobalBalanceOnewayRPC(excludeSelf, comp, method, args...)
}

// BroadcastOnewayRPC 向指定服务中承载当前实体的节点广播单向 RPC；excludeSelf 为 true 时排除源节点。
func (e *EntityBehavior) BroadcastOnewayRPC(excludeSelf bool, service, comp, method string, args ...any) error {
	return rpc.ProxyEntity(e, e.Id()).BroadcastOnewayRPC(excludeSelf, service, comp, method, args...)
}

// GlobalBroadcastOnewayRPC 向所有服务中承载当前实体的节点广播单向 RPC；excludeSelf 为 true 时排除源节点。
func (e *EntityBehavior) GlobalBroadcastOnewayRPC(excludeSelf bool, comp, method string, args ...any) error {
	return rpc.ProxyEntity(e, e.Id()).GlobalBroadcastOnewayRPC(excludeSelf, comp, method, args...)
}

// CliRPC 向当前实体 ID 对应的客户端单播地址发起 RPC。
func (e *EntityBehavior) CliRPC(proc, method string, args ...any) async.Future {
	return rpc.ProxyEntity(e, e.Id()).CliRPC(proc, method, args...)
}

// CliOnewayRPC 向当前实体 ID 对应的客户端单播地址发起单向 RPC。
func (e *EntityBehavior) CliOnewayRPC(proc, method string, args ...any) error {
	return rpc.ProxyEntity(e, e.Id()).CliOnewayRPC(proc, method, args...)
}
