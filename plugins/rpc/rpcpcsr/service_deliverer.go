package rpcpcsr

import (
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/variant"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/plugins/rpcstack"
	"git.golaxy.org/framework/utils/concurrent"
)

// Match 是否匹配
func (p *_ServiceProcessor) Match(ctx service.Context, dst string, callChain rpcstack.CallChain, path string, oneWay bool) bool {
	details := p.dist.GetNodeDetails()

	// 只支持服务域通信
	if !details.InDomain(dst) {
		return false
	}

	if oneWay {
		// 单向请求，支持广播、负载均衡、单播地址
		return details.InBroadcastSubdomain(dst) || details.EqualBroadcastSubdomain(dst) || details.InBalanceSubdomain(dst) || details.EqualBalanceSubdomain(dst) || details.InNodeSubdomain(dst)
	} else {
		// 普通请求，支持负载均衡与单播地址
		return details.InBalanceSubdomain(dst) || details.EqualBalanceSubdomain(dst) || details.InNodeSubdomain(dst)
	}
}

// Request 请求
func (p *_ServiceProcessor) Request(ctx service.Context, dst string, callChain rpcstack.CallChain, path string, args []any) async.AsyncRet {
	ret := concurrent.MakeRespAsyncRet()
	future := concurrent.MakeFuture(p.dist.GetFutures(), nil, ret)

	vargs, err := variant.MakeArrayReadonly(args)
	if err != nil {
		future.Cancel(err)
		return ret.ToAsyncRet()
	}

	msg := &gap.MsgRPCRequest{
		CorrId:    future.Id,
		CallChain: callChain,
		Path:      path,
		Args:      vargs,
	}

	if err = p.dist.SendMsg(dst, msg); err != nil {
		future.Cancel(err)
		return ret.ToAsyncRet()
	}

	log.Debugf(p.servCtx, "rpc request(%d) to dst:%q, path:%q ok", future.Id, dst, path)
	return ret.ToAsyncRet()
}

// Notify 通知
func (p *_ServiceProcessor) Notify(ctx service.Context, dst string, callChain rpcstack.CallChain, path string, args []any) error {
	vargs, err := variant.MakeArrayReadonly(args)
	if err != nil {
		return err
	}

	msg := &gap.MsgOneWayRPC{
		CallChain: callChain,
		Path:      path,
		Args:      vargs,
	}

	if err = p.dist.SendMsg(dst, msg); err != nil {
		return err
	}

	log.Debugf(p.servCtx, "rpc notify to dst:%q, path:%q ok", dst, path)
	return nil
}
