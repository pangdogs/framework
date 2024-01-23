package dserv

import (
	"context"
	"git.golaxy.org/core/service"
	"git.golaxy.org/plugins/gap"
	"git.golaxy.org/plugins/util/concurrent"
)

// GetAddress 获取地址信息
func GetAddress(servCtx service.Context) Address {
	return Using(servCtx).GetAddress()
}

// GetFutures 获取异步模型Future控制器
func GetFutures(servCtx service.Context) concurrent.IFutures {
	return Using(servCtx).GetFutures()
}

// MakeBroadcastAddr 创建服务广播地址
func MakeBroadcastAddr(servCtx service.Context, serviceName string) string {
	return Using(servCtx).MakeBroadcastAddr(serviceName)
}

// MakeBalanceAddr 创建服务负载均衡地址
func MakeBalanceAddr(servCtx service.Context, serviceName string) string {
	return Using(servCtx).MakeBalanceAddr(serviceName)
}

// MakeNodeAddr 创建服务节点地址
func MakeNodeAddr(servCtx service.Context, nodeId string) (string, error) {
	return Using(servCtx).MakeNodeAddr(nodeId)
}

// SendMsg 发送消息
func SendMsg(servCtx service.Context, dst string, msg gap.Msg) error {
	return Using(servCtx).SendMsg(dst, msg)
}

// WatchMsg 监听消息
func WatchMsg(servCtx service.Context, ctx context.Context, handler RecvMsgHandler) IWatcher {
	return Using(servCtx).WatchMsg(ctx, handler)
}
