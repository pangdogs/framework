package framework

import (
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/plugins/rpc/rpcutil"
)

// RPC 向分布式实体目标服务发送RPC
func (c *ComponentBehavior) RPC(service, comp, method string, args ...any) async.AsyncRet {
	return rpcutil.ProxyEntity(c.GetRuntime().Ctx, c.GetEntity().GetId()).RPC(service, comp, method, args...)
}

// BalanceRPC 使用负载均衡模式，向分布式实体目标服务发送RPC
func (c *ComponentBehavior) BalanceRPC(service, comp, method string, args ...any) async.AsyncRet {
	return rpcutil.ProxyEntity(c.GetRuntime().Ctx, c.GetEntity().GetId()).BalanceRPC(service, comp, method, args...)
}

// GlobalBalanceRPC 使用全局负载均衡模式，向分布式实体任意服务发送RPC
func (c *ComponentBehavior) GlobalBalanceRPC(excludeSelf bool, comp, method string, args ...any) async.AsyncRet {
	return rpcutil.ProxyEntity(c.GetRuntime().Ctx, c.GetEntity().GetId()).GlobalBalanceRPC(excludeSelf, comp, method, args...)
}

// OneWayRPC 向分布式实体目标服务发送单向RPC
func (c *ComponentBehavior) OneWayRPC(service, comp, method string, args ...any) error {
	return rpcutil.ProxyEntity(c.GetRuntime().Ctx, c.GetEntity().GetId()).OneWayRPC(service, comp, method, args...)
}

// BalanceOneWayRPC 使用负载均衡模式，向分布式实体目标服务发送单向RPC
func (c *ComponentBehavior) BalanceOneWayRPC(service, comp, method string, args ...any) error {
	return rpcutil.ProxyEntity(c.GetRuntime().Ctx, c.GetEntity().GetId()).BalanceOneWayRPC(service, comp, method, args...)
}

// GlobalBalanceOneWayRPC 使用全局负载均衡模式，向分布式实体任意服务发送单向RPC
func (c *ComponentBehavior) GlobalBalanceOneWayRPC(excludeSelf bool, comp, method string, args ...any) error {
	return rpcutil.ProxyEntity(c.GetRuntime().Ctx, c.GetEntity().GetId()).GlobalBalanceOneWayRPC(excludeSelf, comp, method, args...)
}

// BroadcastOneWayRPC 使用广播模式，向分布式实体目标服务发送单向RPC
func (c *ComponentBehavior) BroadcastOneWayRPC(excludeSelf bool, service, comp, method string, args ...any) error {
	return rpcutil.ProxyEntity(c.GetRuntime().Ctx, c.GetEntity().GetId()).BroadcastOneWayRPC(excludeSelf, service, comp, method, args...)
}

// GlobalBroadcastOneWayRPC 使用全局广播模式，向分布式实体所有服务发送单向RPC
func (c *ComponentBehavior) GlobalBroadcastOneWayRPC(excludeSelf bool, comp, method string, args ...any) error {
	return rpcutil.ProxyEntity(c.GetRuntime().Ctx, c.GetEntity().GetId()).GlobalBroadcastOneWayRPC(excludeSelf, comp, method, args...)
}

// CliRPC 向客户端发送RPC
func (c *ComponentBehavior) CliRPC(method string, args ...any) async.AsyncRet {
	return rpcutil.ProxyEntity(c.GetRuntime().Ctx, c.GetEntity().GetId()).CliRPC(method, args...)
}

// CliRPCToEntity 向客户端实体发送RPC
func (c *ComponentBehavior) CliRPCToEntity(entityId uid.Id, method string, args ...any) async.AsyncRet {
	return rpcutil.ProxyEntity(c.GetRuntime().Ctx, c.GetEntity().GetId()).CliRPCToEntity(entityId, method, args...)
}

// OneWayCliRPC 向客户端发送单向RPC
func (c *ComponentBehavior) OneWayCliRPC(method string, args ...any) error {
	return rpcutil.ProxyEntity(c.GetRuntime().Ctx, c.GetEntity().GetId()).OneWayCliRPC(method, args...)
}

// OneWayCliRPCToEntity 向客户端实体发送单向RPC
func (c *ComponentBehavior) OneWayCliRPCToEntity(entityId uid.Id, method string, args ...any) error {
	return rpcutil.ProxyEntity(c.GetRuntime().Ctx, c.GetEntity().GetId()).OneWayCliRPCToEntity(entityId, method, args...)
}
