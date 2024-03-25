package rpcutil

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/framework/plugins/dserv"
	"git.golaxy.org/framework/plugins/rpc"
	"git.golaxy.org/framework/plugins/rpc/callpath"
)

// GlobalBalanceRPC 使用全局负载均衡模式，向分布式服务发送RPC
func GlobalBalanceRPC(servCtx service.Context, plugin, method string, args ...any) runtime.AsyncRet {
	if servCtx == nil {
		panic(fmt.Errorf("%w: servCtx is nil", core.ErrArgs))
	}

	// 目标地址
	dst := dserv.Using(servCtx).GetNodeDetails().GlobalBalanceAddr

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Service,
		Plugin:   plugin,
		Method:   method,
	}

	return rpc.Using(servCtx).RPC(dst, cp.String(), args...)
}

// GlobalBalanceOneWayRPC 使用全局负载均衡模式，向分布式服务发送单向RPC
func GlobalBalanceOneWayRPC(servCtx service.Context, plugin, method string, args ...any) error {
	if servCtx == nil {
		panic(fmt.Errorf("%w: servCtx is nil", core.ErrArgs))
	}

	// 目标地址
	dst := dserv.Using(servCtx).GetNodeDetails().GlobalBalanceAddr

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Service,
		Plugin:   plugin,
		Method:   method,
	}

	return rpc.Using(servCtx).OneWayRPC(dst, cp.String(), args...)
}

// GlobalBroadcastOneWayRPC 使用全局广播模式，向分布式服务发送单向RPC
func GlobalBroadcastOneWayRPC(servCtx service.Context, plugin, method string, args ...any) error {
	if servCtx == nil {
		panic(fmt.Errorf("%w: servCtx is nil", core.ErrArgs))
	}

	// 目标地址
	dst := dserv.Using(servCtx).GetNodeDetails().GlobalBroadcastAddr

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Service,
		Plugin:   plugin,
		Method:   method,
	}

	return rpc.Using(servCtx).OneWayRPC(dst, cp.String(), args...)
}
