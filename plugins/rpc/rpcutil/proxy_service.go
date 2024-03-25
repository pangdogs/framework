package rpcutil

import (
	"errors"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/uid"
	"git.golaxy.org/framework/plugins/dserv"
	"git.golaxy.org/framework/plugins/rpc"
	"git.golaxy.org/framework/plugins/rpc/callpath"
)

// ProxyService 代理服务
func ProxyService(servCtx service.Context, service string) ServiceProxied {
	return ServiceProxied{
		Context: servCtx,
		Service: service,
	}
}

// ServiceProxied 实体服务，用于向服务发送RPC
type ServiceProxied struct {
	Context service.Context
	Service string
}

// RPC 向分布式服务发送RPC
func (p ServiceProxied) RPC(nodeId uid.Id, plugin, method string, args ...any) runtime.AsyncRet {
	if p.Context == nil {
		panic(errors.New("rpc: setting context is nil"))
	}

	// 目标地址
	dst, err := dserv.Using(p.Context).MakeNodeAddr(nodeId)
	if err != nil {
		return makeErr(err)
	}

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Service,
		Plugin:   plugin,
		Method:   method,
	}

	return rpc.Using(p.Context).RPC(dst, cp.String(), args...)
}

// BalanceRPC 使用负载均衡模式，向分布式服务发送RPC
func (p ServiceProxied) BalanceRPC(plugin, method string, args ...any) runtime.AsyncRet {
	if p.Context == nil {
		panic(errors.New("rpc: setting context is nil"))
	}

	// 目标地址
	dst := dserv.Using(p.Context).MakeBalanceAddr(p.Service)

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Service,
		Plugin:   plugin,
		Method:   method,
	}

	return rpc.Using(p.Context).RPC(dst, cp.String(), args...)
}

// OneWayRPC 向分布式服务发送单向RPC
func (p ServiceProxied) OneWayRPC(nodeId uid.Id, plugin, method string, args ...any) error {
	if p.Context == nil {
		panic(errors.New("rpc: setting context is nil"))
	}

	// 目标地址
	dst, err := dserv.Using(p.Context).MakeNodeAddr(nodeId)
	if err != nil {
		return err
	}

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Service,
		Plugin:   plugin,
		Method:   method,
	}

	return rpc.Using(p.Context).OneWayRPC(dst, cp.String(), args...)
}

// BalanceOneWayRPC 使用负载均衡模式，向分布式服务发送单向RPC
func (p ServiceProxied) BalanceOneWayRPC(plugin, method string, args ...any) error {
	if p.Context == nil {
		panic(errors.New("rpc: setting context is nil"))
	}

	// 目标地址
	dst := dserv.Using(p.Context).MakeBalanceAddr(p.Service)

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Service,
		Plugin:   plugin,
		Method:   method,
	}

	return rpc.Using(p.Context).OneWayRPC(dst, cp.String(), args...)
}

// BroadcastOneWayRPC 使用广播模式，向分布式服务发送单向RPC
func (p ServiceProxied) BroadcastOneWayRPC(plugin, method string, args ...any) error {
	if p.Context == nil {
		panic(errors.New("rpc: setting context is nil"))
	}

	// 目标地址
	dst := dserv.Using(p.Context).MakeBroadcastAddr(p.Service)

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Service,
		Plugin:   plugin,
		Method:   method,
	}

	return rpc.Using(p.Context).OneWayRPC(dst, cp.String(), args...)
}
