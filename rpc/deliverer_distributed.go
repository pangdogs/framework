package rpc

import (
	"kit.golaxy.org/golaxy/runtime"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/plugins/distributed"
	"kit.golaxy.org/plugins/gap"
	"kit.golaxy.org/plugins/gap/variant"
	"kit.golaxy.org/plugins/util/concurrent"
	"strings"
)

// DistributedDeliverer 分布式服务RPC投递器
type DistributedDeliverer struct{}

// Match 是否匹配
func (DistributedDeliverer) Match(ctx service.Context, dst, path string, oneWay bool) bool {
	addr := distributed.Using(ctx).GetAddress()

	if !strings.HasPrefix(dst, addr.Domain) {
		return false
	}

	if !oneWay {
		if !strings.HasPrefix(dst, addr.BalanceSubdomain) && !strings.HasPrefix(dst, addr.NodeSubdomain) {
			return false
		}
	}

	return true
}

// Request 请求
func (DistributedDeliverer) Request(ctx service.Context, dst, path string, args []any) runtime.AsyncRet {
	dist := distributed.Using(ctx)

	ret := concurrent.MakeRespAsyncRet()
	future := concurrent.MakeFuture(dist.GetFutures(), nil, ret)

	vargs, err := variant.MakeArray(args)
	if err != nil {
		future.Cancel(err)
		return ret.Cast()
	}

	msg := &gap.MsgRPCRequest{
		CorrId: future.Id,
		Path:   path,
		Args:   vargs,
	}

	err = dist.SendMsg(dst, msg)
	if err != nil {
		future.Cancel(err)
		return ret.Cast()
	}

	return ret.Cast()
}

// Notify 通知
func (DistributedDeliverer) Notify(ctx service.Context, dst, path string, args []any) error {
	vargs, err := variant.MakeArray(args)
	if err != nil {
		return err
	}

	msg := &gap.MsgOneWayRPC{
		Path: path,
		Args: vargs,
	}

	return distributed.Using(ctx).SendMsg(dst, msg)
}
