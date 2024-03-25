package rpcutil

import (
	"errors"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/uid"
	"git.golaxy.org/framework/net/netpath"
	"git.golaxy.org/framework/plugins/gate"
	"git.golaxy.org/framework/plugins/rpc"
	"git.golaxy.org/framework/plugins/rpc/callpath"
)

// ProxyGroup 代理分组
func ProxyGroup(servCtx service.Context, id uid.Id) GroupProxied {
	return GroupProxied{
		Context: servCtx,
		Id:      id,
	}
}

// GroupProxied 分组代理，用于向分组发送RPC
type GroupProxied struct {
	Context service.Context
	Id      uid.Id
}

// OneWayCliRPC 向分组客户端发送单向RPC
func (p GroupProxied) OneWayCliRPC(method string, args ...any) error {
	return p.OneWayCliRPCToEntity(uid.Nil, method, args...)
}

// OneWayCliRPCToEntity 向分组客户端实体发送单向RPC
func (p GroupProxied) OneWayCliRPCToEntity(entityId uid.Id, method string, args ...any) error {
	if p.Context == nil {
		panic(errors.New("rpc: setting context is nil"))
	}

	// 客户端组播地址
	dst := netpath.Path(gate.CliDetails.PathSeparator, gate.CliDetails.MulticastSubdomain, p.Id.String())

	// 调用路径
	cp := callpath.CallPath{
		Category: callpath.Client,
		EntityId: entityId,
		Method:   method,
	}

	return rpc.Using(p.Context).OneWayRPC(dst, cp.String(), args...)
}
