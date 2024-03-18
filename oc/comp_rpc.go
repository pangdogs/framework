package oc

import (
	"git.golaxy.org/core/runtime"
)

// RPC 向分布式实体目标服务发送RPC
func (c *ComponentBehavior) RPC(service, comp, method string, args ...any) runtime.AsyncRet {
	return ProxyEntity(c.GetServiceCtx(), c.GetEntity().GetId()).RPC(service, comp, method, args...)
}

// BalanceRPC 使用负载均衡模式，向分布式实体目标服务发送RPC
func (c *ComponentBehavior) BalanceRPC(service, comp, method string, args ...any) runtime.AsyncRet {
	return ProxyEntity(c.GetServiceCtx(), c.GetEntity().GetId()).BalanceRPC(service, comp, method, args...)
}

// GlobalBalanceRPC 使用全局负载均衡模式，向分布式实体任意服务发送RPC
func (c *ComponentBehavior) GlobalBalanceRPC(comp, method string, args ...any) runtime.AsyncRet {
	return ProxyEntity(c.GetServiceCtx(), c.GetEntity().GetId()).GlobalBalanceRPC(comp, method, args...)
}

// OneWayRPC 向分布式实体目标服务发送单向RPC
func (c *ComponentBehavior) OneWayRPC(service, comp, method string, args ...any) error {
	return ProxyEntity(c.GetServiceCtx(), c.GetEntity().GetId()).OneWayRPC(service, comp, method, args...)
}

// BalanceOneWayRPC 使用负载均衡模式，向分布式实体目标服务发送单向RPC
func (c *ComponentBehavior) BalanceOneWayRPC(service, comp, method string, args ...any) error {
	return ProxyEntity(c.GetServiceCtx(), c.GetEntity().GetId()).BalanceOneWayRPC(service, comp, method, args...)
}

// GlobalBalanceOneWayRPC 使用全局负载均衡模式，向分布式实体任意服务发送单向RPC
func (c *ComponentBehavior) GlobalBalanceOneWayRPC(comp, method string, args ...any) error {
	return ProxyEntity(c.GetServiceCtx(), c.GetEntity().GetId()).GlobalBroadcastOneWayRPC(comp, method, args...)
}

// BroadcastOneWayRPC 使用广播模式，向分布式实体目标服务发送单向RPC
func (c *ComponentBehavior) BroadcastOneWayRPC(service, comp, method string, args ...any) error {
	return ProxyEntity(c.GetServiceCtx(), c.GetEntity().GetId()).BroadcastOneWayRPC(service, comp, method, args...)
}

// GlobalBroadcastOneWayRPC 使用全局广播模式，向分布式实体所有服务发送单向RPC
func (c *ComponentBehavior) GlobalBroadcastOneWayRPC(comp, method string, args ...any) error {
	return ProxyEntity(c.GetServiceCtx(), c.GetEntity().GetId()).GlobalBroadcastOneWayRPC(comp, method, args...)
}

// CRPC 向客户端发送RPC
func (c *ComponentBehavior) CRPC(method string, args ...any) runtime.AsyncRet {
	return ProxyEntity(c.GetServiceCtx(), c.GetEntity().GetId()).CRPC(method, args...)
}

// OneWayCRPC 向客户端发送单向RPC
func (c *ComponentBehavior) OneWayCRPC(method string, args ...any) error {
	return ProxyEntity(c.GetServiceCtx(), c.GetEntity().GetId()).OneWayCRPC(method, args...)
}
