package rpcutil

import (
	"errors"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/net/netpath"
	"git.golaxy.org/framework/plugins/gate"
	"git.golaxy.org/framework/plugins/rpc"
	"git.golaxy.org/framework/plugins/rpc/callpath"
	"git.golaxy.org/framework/plugins/rpcstack"
)

// ProxyGroup 代理分组
func ProxyGroup(ctx runtime.Context, id uid.Id) GroupProxied {
	return GroupProxied{
		servCtx: service.Current(ctx),
		rtCtx:   ctx,
		id:      id,
	}
}

// ConcurrentProxyGroup 代理分组
func ConcurrentProxyGroup(ctx service.Context, id uid.Id) GroupProxied {
	return GroupProxied{
		servCtx: ctx,
		id:      id,
	}
}

// GroupProxied 分组代理，用于向分组发送RPC
type GroupProxied struct {
	servCtx service.Context
	rtCtx   runtime.Context
	id      uid.Id
}

// GetId 获取分组id
func (p GroupProxied) GetId() uid.Id {
	return p.id
}

// OneWayCliRPC 向分组客户端发送单向RPC
func (p GroupProxied) OneWayCliRPC(method string, args ...any) error {
	return p.OneWayCliRPCToEntity(uid.Nil, method, args...)
}

// OneWayCliRPCToEntity 向分组客户端实体发送单向RPC
func (p GroupProxied) OneWayCliRPCToEntity(entityId uid.Id, method string, args ...any) error {
	if p.servCtx == nil {
		panic(errors.New("rpc: setting servCtx is nil"))
	}

	// 客户端组播地址
	dst := netpath.Path(gate.CliDetails.PathSeparator, gate.CliDetails.MulticastSubdomain, p.id.String())

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

	return rpc.Using(p.servCtx).OneWayRPC(dst, callChain, cp.String(), args...)
}
