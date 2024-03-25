package oc

import (
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/util/uid"
	"git.golaxy.org/framework/plugins/rpc/rpcutil"
)

// RPC 向分布式实体目标服务发送RPC
func (c *ComponentBehavior) RPC(service, comp, method string, args ...any) runtime.AsyncRet {
	return rpcutil.ProxyEntity(c.GetServiceCtx(), c.GetEntity().GetId()).RPC(service, comp, method, args...)
}

// BalanceRPC 使用负载均衡模式，向分布式实体目标服务发送RPC
func (c *ComponentBehavior) BalanceRPC(service, comp, method string, args ...any) runtime.AsyncRet {
	return rpcutil.ProxyEntity(c.GetServiceCtx(), c.GetEntity().GetId()).BalanceRPC(service, comp, method, args...)
}

// GlobalBalanceRPC 使用全局负载均衡模式，向分布式实体任意服务发送RPC
func (c *ComponentBehavior) GlobalBalanceRPC(comp, method string, args ...any) runtime.AsyncRet {
	return rpcutil.ProxyEntity(c.GetServiceCtx(), c.GetEntity().GetId()).GlobalBalanceRPC(comp, method, args...)
}

// OneWayRPC 向分布式实体目标服务发送单向RPC
func (c *ComponentBehavior) OneWayRPC(service, comp, method string, args ...any) error {
	return rpcutil.ProxyEntity(c.GetServiceCtx(), c.GetEntity().GetId()).OneWayRPC(service, comp, method, args...)
}

// BalanceOneWayRPC 使用负载均衡模式，向分布式实体目标服务发送单向RPC
func (c *ComponentBehavior) BalanceOneWayRPC(service, comp, method string, args ...any) error {
	return rpcutil.ProxyEntity(c.GetServiceCtx(), c.GetEntity().GetId()).BalanceOneWayRPC(service, comp, method, args...)
}

// GlobalBalanceOneWayRPC 使用全局负载均衡模式，向分布式实体任意服务发送单向RPC
func (c *ComponentBehavior) GlobalBalanceOneWayRPC(comp, method string, args ...any) error {
	return rpcutil.ProxyEntity(c.GetServiceCtx(), c.GetEntity().GetId()).GlobalBroadcastOneWayRPC(comp, method, args...)
}

// BroadcastOneWayRPC 使用广播模式，向分布式实体目标服务发送单向RPC
func (c *ComponentBehavior) BroadcastOneWayRPC(service, comp, method string, args ...any) error {
	return rpcutil.ProxyEntity(c.GetServiceCtx(), c.GetEntity().GetId()).BroadcastOneWayRPC(service, comp, method, args...)
}

// GlobalBroadcastOneWayRPC 使用全局广播模式，向分布式实体所有服务发送单向RPC
func (c *ComponentBehavior) GlobalBroadcastOneWayRPC(comp, method string, args ...any) error {
	return rpcutil.ProxyEntity(c.GetServiceCtx(), c.GetEntity().GetId()).GlobalBroadcastOneWayRPC(comp, method, args...)
}

// CliRPC 向客户端发送RPC
func (c *ComponentBehavior) CliRPC(method string, args ...any) runtime.AsyncRet {
	return rpcutil.ProxyEntity(c.GetServiceCtx(), c.GetEntity().GetId()).CliRPC(method, args...)
}

// CliRPCToEntity 向客户端实体发送RPC
func (c *ComponentBehavior) CliRPCToEntity(entityId uid.Id, method string, args ...any) runtime.AsyncRet {
	return rpcutil.ProxyEntity(c.GetServiceCtx(), c.GetEntity().GetId()).CliRPCToEntity(entityId, method, args...)
}

// OneWayCliRPC 向客户端发送单向RPC
func (c *ComponentBehavior) OneWayCliRPC(method string, args ...any) error {
	return rpcutil.ProxyEntity(c.GetServiceCtx(), c.GetEntity().GetId()).OneWayCliRPC(method, args...)
}

// OneWayCliRPCToEntity 向客户端实体发送单向RPC
func (c *ComponentBehavior) OneWayCliRPCToEntity(entityId uid.Id, method string, args ...any) error {
	return rpcutil.ProxyEntity(c.GetServiceCtx(), c.GetEntity().GetId()).OneWayCliRPCToEntity(entityId, method, args...)
}
