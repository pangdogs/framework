package rpcutil

import (
	"errors"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/uid"
	"git.golaxy.org/framework/net/netpath"
	"git.golaxy.org/framework/plugins/dentq"
	"git.golaxy.org/framework/plugins/dserv"
	"git.golaxy.org/framework/plugins/gate"
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

// ProxyEntity 代理实体
func ProxyEntity(servCtx service.Context, id uid.Id) EntityProxied {
	return EntityProxied{
		Context: servCtx,
		Id:      id,
	}
}

// EntityProxied 实体代理，用于向实体发送RPC
type EntityProxied struct {
	Context service.Context
	Id      uid.Id
}

// RPC 向分布式实体目标服务发送RPC
func (p EntityProxied) RPC(service, comp, method string, args ...any) runtime.AsyncRet {
	if p.Context == nil {
		panic(errors.New("rpc: setting context is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.Context).GetDistEntity(p.Id)
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
		EntityId:  p.Id,
		Component: comp,
		Method:    method,
	}

	return rpc.Using(p.Context).RPC(distEntity.Nodes[nodeIdx].RemoteAddr, cp.String(), args...)
}

// BalanceRPC 使用负载均衡模式，向分布式实体目标服务发送RPC
func (p EntityProxied) BalanceRPC(service, comp, method string, args ...any) runtime.AsyncRet {
	if p.Context == nil {
		panic(errors.New("rpc: setting context is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.Context).GetDistEntity(p.Id)
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
		EntityId:  p.Id,
		Component: comp,
		Method:    method,
	}

	return rpc.Using(p.Context).RPC(distEntity.Nodes[nodeIdx].BalanceAddr, cp.String(), args...)
}

// GlobalBalanceRPC 使用全局负载均衡模式，向分布式实体任意服务发送RPC
func (p EntityProxied) GlobalBalanceRPC(comp, method string, args ...any) runtime.AsyncRet {
	if p.Context == nil {
		panic(errors.New("rpc: setting context is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.Context).GetDistEntity(p.Id)
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
		EntityId:  p.Id,
		Component: comp,
		Method:    method,
	}

	return rpc.Using(p.Context).RPC(dst, cp.String(), args...)
}

// OneWayRPC 向分布式实体目标服务发送单向RPC
func (p EntityProxied) OneWayRPC(service, comp, method string, args ...any) error {
	if p.Context == nil {
		panic(errors.New("rpc: setting context is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.Context).GetDistEntity(p.Id)
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
		EntityId:  p.Id,
		Component: comp,
		Method:    method,
	}

	return rpc.Using(p.Context).OneWayRPC(distEntity.Nodes[nodeIdx].RemoteAddr, cp.String(), args...)
}

// BalanceOneWayRPC 使用负载均衡模式，向分布式实体目标服务发送单向RPC
func (p EntityProxied) BalanceOneWayRPC(service, comp, method string, args ...any) error {
	if p.Context == nil {
		panic(errors.New("rpc: setting context is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.Context).GetDistEntity(p.Id)
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
		EntityId:  p.Id,
		Component: comp,
		Method:    method,
	}

	return rpc.Using(p.Context).OneWayRPC(distEntity.Nodes[nodeIdx].BalanceAddr, cp.String(), args...)
}

// GlobalBalanceOneWayRPC 使用全局负载均衡模式，向分布式实体任意服务发送单向RPC
func (p EntityProxied) GlobalBalanceOneWayRPC(comp, method string, args ...any) error {
	if p.Context == nil {
		panic(errors.New("rpc: setting context is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.Context).GetDistEntity(p.Id)
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
		EntityId:  p.Id,
		Component: comp,
		Method:    method,
	}

	return rpc.Using(p.Context).OneWayRPC(dst, cp.String(), args...)
}

// BroadcastOneWayRPC 使用广播模式，向分布式实体目标服务发送单向RPC
func (p EntityProxied) BroadcastOneWayRPC(service, comp, method string, args ...any) error {
	if p.Context == nil {
		panic(errors.New("rpc: setting context is nil"))
	}

	// 查询分布式实体信息
	distEntity, ok := dentq.Using(p.Context).GetDistEntity(p.Id)
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
		EntityId:  p.Id,
		Component: comp,
		Method:    method,
	}

	return rpc.Using(p.Context).OneWayRPC(distEntity.Nodes[nodeIdx].BroadcastAddr, cp.String(), args...)
}

// GlobalBroadcastOneWayRPC 使用全局广播模式，向分布式实体所有服务发送单向RPC
func (p EntityProxied) GlobalBroadcastOneWayRPC(comp, method string, args ...any) error {
	if p.Context == nil {
		panic(errors.New("rpc: setting context is nil"))
	}

	// 全局广播地址
	dst := dserv.Using(p.Context).GetNodeDetails().GlobalBroadcastAddr

	// 调用路径
	cp := callpath.CallPath{
		Category:  callpath.Entity,
		EntityId:  p.Id,
		Component: comp,
		Method:    method,
	}

	return rpc.Using(p.Context).OneWayRPC(dst, cp.String(), args...)
}

// CliRPC 向客户端发送RPC
func (p EntityProxied) CliRPC(method string, args ...any) runtime.AsyncRet {
	return p.CliRPCToEntity(uid.Nil, method, args...)
}

// CliRPCToEntity 向客户端实体发送RPC
func (p EntityProxied) CliRPCToEntity(entityId uid.Id, method string, args ...any) runtime.AsyncRet {
	if p.Context == nil {
		panic(errors.New("rpc: setting context is nil"))
	}

	// 客户端地址
	dst := netpath.Path(gate.CliDetails.PathSeparator, gate.CliDetails.NodeSubdomain, p.Id.String())

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Client,
		EntityId: entityId,
		Method:   method,
	}

	return rpc.Using(p.Context).RPC(dst, cp.String(), args...)
}

// OneWayCliRPC 向客户端发送单向RPC
func (p EntityProxied) OneWayCliRPC(method string, args ...any) error {
	return p.OneWayCliRPCToEntity(uid.Nil, method, args...)
}

// OneWayCliRPCToEntity 向客户端实体发送单向RPC
func (p EntityProxied) OneWayCliRPCToEntity(entityId uid.Id, method string, args ...any) error {
	if p.Context == nil {
		panic(errors.New("rpc: setting context is nil"))
	}

	// 客户端地址
	dst := netpath.Path(gate.CliDetails.PathSeparator, gate.CliDetails.NodeSubdomain, p.Id.String())

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Client,
		EntityId: entityId,
		Method:   method,
	}

	return rpc.Using(p.Context).OneWayRPC(dst, cp.String(), args...)
}
