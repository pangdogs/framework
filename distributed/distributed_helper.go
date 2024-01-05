package distributed

import (
	"golang.org/x/net/context"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/plugins/gap"
	"kit.golaxy.org/plugins/util/concurrent"
)

// GetAddress 获取地址信息
func GetAddress(servCtx service.Context) Address {
	return Using(servCtx).GetAddress()
}

// GetFutures 获取异步模型Future控制器
func GetFutures(servCtx service.Context) concurrent.IFutures {
	return Using(servCtx).GetFutures()
}

// MakeServiceBroadcastAddr 创建服务广播地址
func MakeServiceBroadcastAddr(servCtx service.Context, serviceName string) string {
	return Using(servCtx).MakeServiceBroadcastAddr(serviceName)
}

// MakeServiceBalanceAddr 创建服务负载均衡地址
func MakeServiceBalanceAddr(servCtx service.Context, serviceName string) string {
	return Using(servCtx).MakeServiceBalanceAddr(serviceName)
}

// MakeServiceNodeAddr 创建服务节点地址
func MakeServiceNodeAddr(servCtx service.Context, serviceName, nodeId string) (string, error) {
	return Using(servCtx).MakeServiceNodeAddr(serviceName, nodeId)
}

// SendMsg 发送消息
func SendMsg(servCtx service.Context, dst string, msg gap.Msg) error {
	return Using(servCtx).SendMsg(dst, msg)
}

// WatchMsg 监听消息
func WatchMsg(servCtx service.Context, ctx context.Context, handler RecvMsgHandler) Watcher {
	return Using(servCtx).WatchMsg(ctx, handler)
}