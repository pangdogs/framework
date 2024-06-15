package rpcutil

import (
	"errors"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/plugins/rpc"
	"git.golaxy.org/framework/plugins/rpc/callpath"
	"git.golaxy.org/framework/plugins/rpcstack"
)

// ProxyGroup 代理分组
func ProxyGroup(ctx runtime.CurrentContextProvider, addr string) GroupProxied {
	return GroupProxied{
		servCtx: service.Current(ctx),
		rtCtx:   runtime.Current(ctx),
		addr:    addr,
	}
}

// ConcurrentProxyGroup 代理分组
func ConcurrentProxyGroup(ctx service.Context, addr string) GroupProxied {
	return GroupProxied{
		servCtx: ctx,
		addr:    addr,
	}
}

// GroupProxied 分组代理，用于向分组发送RPC
type GroupProxied struct {
	servCtx service.Context
	rtCtx   runtime.Context
	addr    string
}

// GetAddr 获取分组地址
func (p GroupProxied) GetAddr() string {
	return p.addr
}

// OneWayCliRPC 向分组发送单向RPC
func (p GroupProxied) OneWayCliRPC(method string, args ...any) error {
	return p.OneWayCliRPCToEntity(uid.Nil, method, args...)
}

// OneWayCliRPCToEntity 向分组发送单向RPC
func (p GroupProxied) OneWayCliRPCToEntity(entityId uid.Id, method string, args ...any) error {
	if p.servCtx == nil {
		panic(errors.New("rpc: setting servCtx is nil"))
	}

	// 调用链
	callChain := rpcstack.EmptyCallChain
	if p.rtCtx != nil {
		callChain = rpcstack.Using(p.rtCtx).CallChain()
	}

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Client,
		EntityId: entityId,
		Method:   method,
	}

	return rpc.Using(p.servCtx).OneWayRPC(p.addr, callChain, cp.String(), args...)
}
