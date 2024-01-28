package oc

import (
	"errors"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/framework/plugins/dentq"
	"git.golaxy.org/framework/plugins/dserv"
	"git.golaxy.org/framework/plugins/rpc"
	"git.golaxy.org/framework/plugins/rpc/callpath"
	"github.com/elliotchance/pie/v2"
	"math/rand"
)

var (
	ErrDistEntityNotFound     = errors.New("rpc: distributed entity not found")
	ErrDistEntityNodeNotFound = errors.New("rpc: distributed entity node not found")
)

func makeErr(err error) (asyncRet chan runtime.Ret) {
	asyncRet = make(chan runtime.Ret, 1)
	asyncRet <- runtime.MakeRet(nil, err)
	close(asyncRet)
	return
}

// RPC 向分布式实体目标服务发送RPC
func (c *ComponentBehavior) RPC(service, comp, method string, args ...any) runtime.AsyncRet {
	// 查询分布式实体信息
	distEntity, ok := dentq.Using(c.GetServCtx()).GetDistEntity(c.GetEntity().GetId())
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
		Category:  callpath.Entity,
		EntityId:  c.GetEntity().GetId().String(),
		Component: comp,
		Method:    method,
	}

	return rpc.RPC(c.GetServCtx(), distEntity.Nodes[nodeIdx].RemoteAddr, cp.String(), args...)
}

// BalanceRPC 使用负载均衡模式，向分布式实体目标服务发送RPC
func (c *ComponentBehavior) BalanceRPC(service, comp, method string, args ...any) runtime.AsyncRet {
	// 查询分布式实体信息
	distEntity, ok := dentq.Using(c.GetServCtx()).GetDistEntity(c.GetEntity().GetId())
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
		Category:  callpath.Entity,
		EntityId:  c.GetEntity().GetId().String(),
		Component: comp,
		Method:    method,
	}

	return rpc.RPC(c.GetServCtx(), distEntity.Nodes[nodeIdx].BalanceAddr, cp.String(), args...)
}

// GlobalBalanceRPC 使用全局负载均衡模式，向分布式实体任意服务发送RPC
func (c *ComponentBehavior) GlobalBalanceRPC(comp, method string, args ...any) runtime.AsyncRet {
	// 查询分布式实体信息
	distEntity, ok := dentq.Using(c.GetServCtx()).GetDistEntity(c.GetEntity().GetId())
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
		Category:  callpath.Entity,
		EntityId:  c.GetEntity().GetId().String(),
		Component: comp,
		Method:    method,
	}

	return rpc.RPC(c.GetServCtx(), dst, cp.String(), args...)
}

// OneWayRPC 向分布式实体目标服务发送单向RPC
func (c *ComponentBehavior) OneWayRPC(service, comp, method string, args ...any) error {
	// 查询分布式实体信息
	distEntity, ok := dentq.Using(c.GetServCtx()).GetDistEntity(c.GetEntity().GetId())
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
		Category:  callpath.Entity,
		EntityId:  c.GetEntity().GetId().String(),
		Component: comp,
		Method:    method,
	}

	return rpc.OneWayRPC(c.GetServCtx(), distEntity.Nodes[nodeIdx].RemoteAddr, cp.String(), args...)
}

// BalanceOneWayRPC 使用负载均衡模式，向分布式实体目标服务发送单向RPC
func (c *ComponentBehavior) BalanceOneWayRPC(service, comp, method string, args ...any) error {
	// 查询分布式实体信息
	distEntity, ok := dentq.Using(c.GetServCtx()).GetDistEntity(c.GetEntity().GetId())
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
		Category:  callpath.Entity,
		EntityId:  c.GetEntity().GetId().String(),
		Component: comp,
		Method:    method,
	}

	return rpc.OneWayRPC(c.GetServCtx(), distEntity.Nodes[nodeIdx].BalanceAddr, cp.String(), args...)
}

// GlobalBalanceOneWayRPC 使用全局负载均衡模式，向分布式实体任意服务发送单向RPC
func (c *ComponentBehavior) GlobalBalanceOneWayRPC(comp, method string, args ...any) error {
	// 查询分布式实体信息
	distEntity, ok := dentq.Using(c.GetServCtx()).GetDistEntity(c.GetEntity().GetId())
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
		Category:  callpath.Entity,
		EntityId:  c.GetEntity().GetId().String(),
		Component: comp,
		Method:    method,
	}

	return rpc.OneWayRPC(c.GetServCtx(), dst, cp.String(), args...)
}

// BroadcastOneWayRPC 使用广播模式，向分布式实体目标服务发送单向RPC
func (c *ComponentBehavior) BroadcastOneWayRPC(service, comp, method string, args ...any) error {
	// 查询分布式实体信息
	distEntity, ok := dentq.Using(c.GetServCtx()).GetDistEntity(c.GetEntity().GetId())
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
		Category:  callpath.Entity,
		EntityId:  c.GetEntity().GetId().String(),
		Component: comp,
		Method:    method,
	}

	return rpc.OneWayRPC(c.GetServCtx(), distEntity.Nodes[nodeIdx].BroadcastAddr, cp.String(), args...)
}

// GlobalBroadcastOneWayRPC 使用全局广播模式，向分布式实体所有服务发送单向RPC
func (c *ComponentBehavior) GlobalBroadcastOneWayRPC(comp, method string, args ...any) error {
	// 全局广播地址
	dst := dserv.Using(c.GetServCtx()).GetAddress().GlobalBroadcastAddr

	// 调用路径
	cp := callpath.CallPath{
		Category:  callpath.Entity,
		EntityId:  c.GetEntity().GetId().String(),
		Component: comp,
		Method:    method,
	}

	return rpc.OneWayRPC(c.GetServCtx(), dst, cp.String(), args...)
}
