package framework

import (
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/util/uid"
	"git.golaxy.org/framework/plugins/rpc/rpcutil"
)

// RPC 向分布式实体目标服务发送RPC
func (e *EntityBehavior) RPC(service, comp, method string, args ...any) runtime.AsyncRet {
	return rpcutil.ProxyEntity(e.GetRuntime().Ctx, e.GetId()).RPC(service, comp, method, args...)
}

// BalanceRPC 使用负载均衡模式，向分布式实体目标服务发送RPC
func (e *EntityBehavior) BalanceRPC(service, comp, method string, args ...any) runtime.AsyncRet {
	return rpcutil.ProxyEntity(e.GetRuntime().Ctx, e.GetId()).BalanceRPC(service, comp, method, args...)
}

// GlobalBalanceRPC 使用全局负载均衡模式，向分布式实体任意服务发送RPC
func (e *EntityBehavior) GlobalBalanceRPC(comp, method string, args ...any) runtime.AsyncRet {
	return rpcutil.ProxyEntity(e.GetRuntime().Ctx, e.GetId()).GlobalBalanceRPC(comp, method, args...)
}

// OneWayRPC 向分布式实体目标服务发送单向RPC
func (e *EntityBehavior) OneWayRPC(service, comp, method string, args ...any) error {
	return rpcutil.ProxyEntity(e.GetRuntime().Ctx, e.GetId()).OneWayRPC(service, comp, method, args...)
}

// BalanceOneWayRPC 使用负载均衡模式，向分布式实体目标服务发送单向RPC
func (e *EntityBehavior) BalanceOneWayRPC(service, comp, method string, args ...any) error {
	return rpcutil.ProxyEntity(e.GetRuntime().Ctx, e.GetId()).BalanceOneWayRPC(service, comp, method, args...)
}

// GlobalBalanceOneWayRPC 使用全局负载均衡模式，向分布式实体任意服务发送单向RPC
func (e *EntityBehavior) GlobalBalanceOneWayRPC(comp, method string, args ...any) error {
	return rpcutil.ProxyEntity(e.GetRuntime().Ctx, e.GetId()).GlobalBalanceOneWayRPC(comp, method, args...)
}

// BroadcastOneWayRPC 使用广播模式，向分布式实体目标服务发送单向RPC
func (e *EntityBehavior) BroadcastOneWayRPC(excludeSelf bool, service, comp, method string, args ...any) error {
	return rpcutil.ProxyEntity(e.GetRuntime().Ctx, e.GetId()).BroadcastOneWayRPC(excludeSelf, service, comp, method, args...)
}

// GlobalBroadcastOneWayRPC 使用全局广播模式，向分布式实体所有服务发送单向RPC
func (e *EntityBehavior) GlobalBroadcastOneWayRPC(excludeSelf bool, comp, method string, args ...any) error {
	return rpcutil.ProxyEntity(e.GetRuntime().Ctx, e.GetId()).GlobalBroadcastOneWayRPC(excludeSelf, comp, method, args...)
}

// CliRPC 向客户端发送RPC
func (e *EntityBehavior) CliRPC(method string, args ...any) runtime.AsyncRet {
	return rpcutil.ProxyEntity(e.GetRuntime().Ctx, e.GetId()).CliRPC(method, args...)
}

// CliRPCToEntity 向客户端实体发送RPC
func (e *EntityBehavior) CliRPCToEntity(entityId uid.Id, method string, args ...any) runtime.AsyncRet {
	return rpcutil.ProxyEntity(e.GetRuntime().Ctx, e.GetId()).CliRPCToEntity(entityId, method, args...)
}

// OneWayCliRPC 向客户端发送单向RPC
func (e *EntityBehavior) OneWayCliRPC(method string, args ...any) error {
	return rpcutil.ProxyEntity(e.GetRuntime().Ctx, e.GetId()).OneWayCliRPC(method, args...)
}

// OneWayCliRPCToEntity 向客户端实体发送单向RPC
func (e *EntityBehavior) OneWayCliRPCToEntity(entityId uid.Id, method string, args ...any) error {
	return rpcutil.ProxyEntity(e.GetRuntime().Ctx, e.GetId()).OneWayCliRPCToEntity(entityId, method, args...)
}
