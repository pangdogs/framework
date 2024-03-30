package rpcutil

import (
	"errors"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/uid"
	"git.golaxy.org/framework/plugins/dentq"
	"git.golaxy.org/framework/plugins/dserv"
	"git.golaxy.org/framework/plugins/rpc"
	"git.golaxy.org/framework/plugins/rpc/callpath"
	"github.com/elliotchance/pie/v2"
	"math/rand"
)

// ProxyRuntime 代理运行时
func ProxyRuntime(ctx service.Context, entityId uid.Id) RuntimeProxied {
	return RuntimeProxied{
		servCtx:  ctx,
		entityId: entityId,
	}
}

// RuntimeProxied 运行时代理，用于向实体的运行时发送RPC
type RuntimeProxied struct {
	servCtx  service.Context
	entityId uid.Id
}

// GetEntityId 获取实体id
func (p RuntimeProxied) GetEntityId() uid.Id {
	return p.entityId
}

// RPC 向分布式实体目标服务的运行时发送RPC
func (p RuntimeProxied) RPC(service, plugin, method string, args ...any) runtime.AsyncRet {
	if p.servCtx == nil {
		panic(errors.New("rpc: setting servCtx is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.servCtx).GetDistEntity(p.entityId)
	if !ok {
		return makeErr(ErrDistEntityNotFound)
	}

	// 查询分布式实体目标服务节点
	nodeIdx := pie.FindFirstUsing(distEntity.Nodes, func(node dentq.Node) bool {
		return node.Service == service
	})
	if nodeIdx < 0 {
		return makeErr(ErrDistEntityNodeNotFound)
	}

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Entity,
		EntityId: p.entityId,
		Plugin:   plugin,
		Method:   method,
	}

	return rpc.Using(p.servCtx).RPC(distEntity.Nodes[nodeIdx].RemoteAddr, cp.String(), args...)
}

// BalanceRPC 使用负载均衡模式，向分布式实体目标服务的运行时发送RPC
func (p RuntimeProxied) BalanceRPC(service, plugin, method string, args ...any) runtime.AsyncRet {
	if p.servCtx == nil {
		panic(errors.New("rpc: setting servCtx is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.servCtx).GetDistEntity(p.entityId)
	if !ok {
		return makeErr(ErrDistEntityNotFound)
	}

	// 查询分布式实体目标服务节点
	nodeIdx := pie.FindFirstUsing(distEntity.Nodes, func(node dentq.Node) bool {
		return node.Service == service
	})
	if nodeIdx < 0 {
		return makeErr(ErrDistEntityNodeNotFound)
	}

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Entity,
		EntityId: p.entityId,
		Plugin:   plugin,
		Method:   method,
	}

	return rpc.Using(p.servCtx).RPC(distEntity.Nodes[nodeIdx].BalanceAddr, cp.String(), args...)
}

// GlobalBalanceRPC 使用全局负载均衡模式，向分布式实体任意服务的运行时发送RPC
func (p RuntimeProxied) GlobalBalanceRPC(plugin, method string, args ...any) runtime.AsyncRet {
	if p.servCtx == nil {
		panic(errors.New("rpc: setting servCtx is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.servCtx).GetDistEntity(p.entityId)
	if !ok {
		return makeErr(ErrDistEntityNotFound)
	}

	// 随机获取服务地址
	if len(distEntity.Nodes) <= 0 {
		return makeErr(ErrDistEntityNodeNotFound)
	}
	dst := distEntity.Nodes[rand.Intn(len(distEntity.Nodes))].RemoteAddr

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Entity,
		EntityId: p.entityId,
		Plugin:   plugin,
		Method:   method,
	}

	return rpc.Using(p.servCtx).RPC(dst, cp.String(), args...)
}

// OneWayRPC 向分布式实体目标服务的运行时发送单向RPC
func (p RuntimeProxied) OneWayRPC(service, plugin, method string, args ...any) error {
	if p.servCtx == nil {
		panic(errors.New("rpc: setting servCtx is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.servCtx).GetDistEntity(p.entityId)
	if !ok {
		return ErrDistEntityNotFound
	}

	// 查询分布式实体目标服务节点
	nodeIdx := pie.FindFirstUsing(distEntity.Nodes, func(node dentq.Node) bool {
		return node.Service == service
	})
	if nodeIdx < 0 {
		return ErrDistEntityNodeNotFound
	}

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Entity,
		EntityId: p.entityId,
		Plugin:   plugin,
		Method:   method,
	}

	return rpc.Using(p.servCtx).OneWayRPC(distEntity.Nodes[nodeIdx].RemoteAddr, cp.String(), args...)
}

// BalanceOneWayRPC 使用负载均衡模式，向分布式实体目标服务的运行时发送单向RPC
func (p RuntimeProxied) BalanceOneWayRPC(service, plugin, method string, args ...any) error {
	if p.servCtx == nil {
		panic(errors.New("rpc: setting servCtx is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.servCtx).GetDistEntity(p.entityId)
	if !ok {
		return ErrDistEntityNotFound
	}

	// 查询分布式实体目标服务节点
	nodeIdx := pie.FindFirstUsing(distEntity.Nodes, func(node dentq.Node) bool {
		return node.Service == service
	})
	if nodeIdx < 0 {
		return ErrDistEntityNodeNotFound
	}

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Entity,
		EntityId: p.entityId,
		Plugin:   plugin,
		Method:   method,
	}

	return rpc.Using(p.servCtx).OneWayRPC(distEntity.Nodes[nodeIdx].BalanceAddr, cp.String(), args...)
}

// GlobalBalanceOneWayRPC 使用全局负载均衡模式，向分布式实体任意服务的运行时发送单向RPC
func (p RuntimeProxied) GlobalBalanceOneWayRPC(plugin, method string, args ...any) error {
	if p.servCtx == nil {
		panic(errors.New("rpc: setting servCtx is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.servCtx).GetDistEntity(p.entityId)
	if !ok {
		return ErrDistEntityNotFound
	}

	// 随机获取服务地址
	if len(distEntity.Nodes) <= 0 {
		return ErrDistEntityNodeNotFound
	}
	dst := distEntity.Nodes[rand.Intn(len(distEntity.Nodes))].RemoteAddr

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Entity,
		EntityId: p.entityId,
		Plugin:   plugin,
		Method:   method,
	}

	return rpc.Using(p.servCtx).OneWayRPC(dst, cp.String(), args...)
}

// BroadcastOneWayRPC 使用广播模式，向分布式实体目标服务的运行时发送单向RPC
func (p RuntimeProxied) BroadcastOneWayRPC(service, plugin, method string, args ...any) error {
	if p.servCtx == nil {
		panic(errors.New("rpc: setting servCtx is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.servCtx).GetDistEntity(p.entityId)
	if !ok {
		return ErrDistEntityNotFound
	}

	// 查询分布式实体目标服务节点
	nodeIdx := pie.FindFirstUsing(distEntity.Nodes, func(node dentq.Node) bool {
		return node.Service == service
	})
	if nodeIdx < 0 {
		return ErrDistEntityNodeNotFound
	}

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Entity,
		EntityId: p.entityId,
		Plugin:   plugin,
		Method:   method,
	}

	return rpc.Using(p.servCtx).OneWayRPC(distEntity.Nodes[nodeIdx].BroadcastAddr, cp.String(), args...)
}

// GlobalBroadcastOneWayRPC 使用全局广播模式，向分布式实体所有服务的运行时发送单向RPC
func (p RuntimeProxied) GlobalBroadcastOneWayRPC(plugin, method string, args ...any) error {
	if p.servCtx == nil {
		panic(errors.New("rpc: setting servCtx is nil"))
	}

	// 全局广播地址
	dst := dserv.Using(p.servCtx).GetNodeDetails().GlobalBroadcastAddr

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Entity,
		EntityId: p.entityId,
		Plugin:   plugin,
		Method:   method,
	}

	return rpc.Using(p.servCtx).OneWayRPC(dst, cp.String(), args...)
}
