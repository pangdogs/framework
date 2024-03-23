package processor

import (
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/types"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/variant"
	"git.golaxy.org/framework/plugins/dserv"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/util/concurrent"
)

// NewServiceDeliverer 创建分布式服务间的RPC投递器
func NewServiceDeliverer() IDeliverer {
	return &_ServiceDeliverer{}
}

// _ServiceDeliverer 分布式服务间的RPC投递器
type _ServiceDeliverer struct {
	servCtx service.Context
	dist    dserv.IDistService
}

// Init 初始化
func (d *_ServiceDeliverer) Init(ctx service.Context) {
	d.servCtx = ctx
	d.dist = dserv.Using(ctx)

	log.Debugf(d.servCtx, "rpc deliverer %q started", types.AnyFullName(*d))
}

// Shut 结束
func (d *_ServiceDeliverer) Shut(ctx service.Context) {
	log.Debugf(d.servCtx, "rpc deliverer %q stopped", types.AnyFullName(*d))
}

// Match 是否匹配
func (d *_ServiceDeliverer) Match(ctx service.Context, dst, path string, oneWay bool) bool {
	details := d.dist.GetNodeDetails()

	// 只支持服务域通信
	if !details.InDomain(dst) {
		return false
	}

	if oneWay {
		// 单向请求，支持广播、负载均衡、单播地址
		return details.SameBroadcastSubdomain(dst) || details.SameBalanceSubdomain(dst) || details.InNodeSubdomain(dst)
	} else {
		// 普通请求，支持负载均衡与单播地址
		return details.SameBalanceSubdomain(dst) || details.InNodeSubdomain(dst)
	}
}

// Request 请求
func (d *_ServiceDeliverer) Request(ctx service.Context, dst, path string, args []any) runtime.AsyncRet {
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

	log.Debugf(d.servCtx, "rpc request(%d) to dst:%q, path:%q ok", future.Id, dst, path)

	return ret.CastAsyncRet()
}

// Notify 通知
func (d *_ServiceDeliverer) Notify(ctx service.Context, dst, path string, args []any) error {
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
