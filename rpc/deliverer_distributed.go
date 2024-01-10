package rpc

import (
	"kit.golaxy.org/golaxy/runtime"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util/types"
	"kit.golaxy.org/plugins/distributed"
	"kit.golaxy.org/plugins/gap"
	"kit.golaxy.org/plugins/gap/variant"
	"kit.golaxy.org/plugins/log"
	"kit.golaxy.org/plugins/util/concurrent"
	"strings"
)

// DistributedDeliverer 分布式服务的RPC投递器
type DistributedDeliverer struct {
	servCtx service.Context
	dist    distributed.IDistributed
}

// Init 初始化
func (d *DistributedDeliverer) Init(ctx service.Context) {
	d.servCtx = ctx
	d.dist = distributed.Using(ctx)

	log.Debugf(d.servCtx, "rpc deliverer %q started", types.AnyFullName(*d))
}

// Shut 结束
func (d *DistributedDeliverer) Shut(ctx service.Context) {
	log.Debugf(d.servCtx, "rpc deliverer %q stopped", types.AnyFullName(*d))
}

// Match 是否匹配
func (d *DistributedDeliverer) Match(ctx service.Context, dst, path string, oneWay bool) bool {
	addr := d.dist.GetAddress()

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
func (d *DistributedDeliverer) Request(ctx service.Context, dst, path string, args []any) runtime.AsyncRet {
	ret := concurrent.MakeRespAsyncRet()
	future := concurrent.MakeFuture(d.dist.GetFutures(), nil, ret)

	vargs, err := variant.MakeArray(args)
	if err != nil {
		future.Cancel(err)
		return ret.CastAsyncRet()
	}

	msg := &gap.MsgRPCRequest{
		CorrId: future.Id,
		Path:   path,
		Args:   vargs,
	}

	if err = d.dist.SendMsg(dst, msg); err != nil {
		future.Cancel(err)
		return ret.CastAsyncRet()
	}

	log.Debugf(d.servCtx, "rpc(%d) request to dst:%q, path:%q ok", future.Id, dst, path)

	return ret.CastAsyncRet()
}

// Notify 通知
func (d *DistributedDeliverer) Notify(ctx service.Context, dst, path string, args []any) error {
	vargs, err := variant.MakeArray(args)
	if err != nil {
		return err
	}

	msg := &gap.MsgOneWayRPC{
		Path: path,
		Args: vargs,
	}

	if err = d.dist.SendMsg(dst, msg); err != nil {
		return err
	}

	log.Debugf(d.servCtx, "rpc notify to dst:%q, path:%q ok", dst, path)

	return nil
}
