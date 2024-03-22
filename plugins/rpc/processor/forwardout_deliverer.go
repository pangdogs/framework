package processor

import (
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/types"
	"git.golaxy.org/core/util/uid"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/variant"
	"git.golaxy.org/framework/net/netpath"
	"git.golaxy.org/framework/plugins/dentq"
	"git.golaxy.org/framework/plugins/dserv"
	"git.golaxy.org/framework/plugins/gate"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/util/binaryutil"
	"git.golaxy.org/framework/util/concurrent"
	"github.com/elliotchance/pie/v2"
)

// NewForwardOutDeliverer RPC外转投递器，用于S->C的通信
func NewForwardOutDeliverer(gate string) IDeliverer {
	return &_ForwardOutDeliverer{
		gate: gate,
	}
}

// _ForwardOutDeliverer RPC外转投递器，用于S->C的通信
type _ForwardOutDeliverer struct {
	servCtx         service.Context
	dist            dserv.IDistService
	dentq           dentq.IDistEntityQuerier
	gate            string
	multicastBCAddr string
}

// Init 初始化
func (d *_ForwardOutDeliverer) Init(ctx service.Context) {
	d.servCtx = ctx
	d.dist = dserv.Using(ctx)
	d.dentq = dentq.Using(ctx)
	d.multicastBCAddr = d.dist.MakeBroadcastAddr(d.gate)

	log.Debugf(d.servCtx, "rpc deliverer %q started", types.AnyFullName(*d))
}

// Shut 结束
func (d *_ForwardOutDeliverer) Shut(ctx service.Context) {
	log.Debugf(d.servCtx, "rpc deliverer %q stopped", types.AnyFullName(*d))
}

// Match 是否匹配
func (d *_ForwardOutDeliverer) Match(ctx service.Context, dst, path string, oneWay bool) bool {
	// 只支持客户端域通信
	if !gate.ClientAddressDetails.InDomain(dst) {
		return false
	}

	if oneWay {
		// 普通请求，支持单播地址
		return gate.ClientAddressDetails.InNodeSubdomain(dst)
	} else {
		// 单向请求，支持组播、单播地址
		return gate.ClientAddressDetails.InNodeSubdomain(dst) || gate.ClientAddressDetails.InMulticastSubdomain(dst)
	}
}

// Request 请求
func (d *_ForwardOutDeliverer) Request(ctx service.Context, dst, path string, args []any) runtime.AsyncRet {
	ret := concurrent.MakeRespAsyncRet()
	future := concurrent.MakeFuture(d.dist.GetFutures(), nil, ret)

	forwardAddr, err := d.getDistEntityForwardAddr(uid.From(netpath.Base(gate.ClientAddressDetails.PathSeparator, dst)))
	if err != nil {
		future.Cancel(err)
		return ret.CastAsyncRet()
	}

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

	bs := binaryutil.MakeRecycleBytes(binaryutil.BytesPool.Get(msg.Size()))
	defer bs.Release()

	if _, err = msg.Read(bs.Data()); err != nil {
		future.Cancel(err)
		return ret.CastAsyncRet()
	}

	forwardMsg := &gap.MsgForward{
		Dst:       dst,
		CorrId:    msg.CorrId,
		TransId:   msg.MsgId(),
		TransData: bs.Data(),
	}

	if err = d.dist.SendMsg(forwardAddr, forwardMsg); err != nil {
		future.Cancel(err)
		return ret.CastAsyncRet()
	}

	log.Debugf(d.servCtx, "rpc request(%d) forwarding to dst:%q, path:%q ok", future.Id, forwardAddr, path)

	return ret.CastAsyncRet()
}

// Notify 通知
func (d *_ForwardOutDeliverer) Notify(ctx service.Context, dst, path string, args []any) error {
	forwardAddr, err := d.getForwardAddr(dst)
	if err != nil {
		return err
	}

	vargs, err := variant.MakeArray(args)
	if err != nil {
		return err
	}

	msg := &gap.MsgOneWayRPC{
		Path: path,
		Args: vargs,
	}

	bs := binaryutil.MakeRecycleBytes(binaryutil.BytesPool.Get(msg.Size()))
	defer bs.Release()

	if _, err = msg.Read(bs.Data()); err != nil {
		return err
	}

	forwardMsg := &gap.MsgForward{
		Dst:       dst,
		TransId:   msg.MsgId(),
		TransData: bs.Data(),
	}

	if err = d.dist.SendMsg(forwardAddr, forwardMsg); err != nil {
		return err
	}

	log.Debugf(d.servCtx, "rpc notify forwarding to dst:%q, path:%q ok", forwardAddr, path)

	return nil
}

func (d *_ForwardOutDeliverer) getForwardAddr(dst string) (string, error) {
	if gate.ClientAddressDetails.InNodeSubdomain(dst) {
		// 目标为单播地址，查询实体的通信中转服务地址
		return d.getDistEntityForwardAddr(uid.From(netpath.Base(gate.ClientAddressDetails.PathSeparator, dst)))

	} else if gate.ClientAddressDetails.InMulticastSubdomain(dst) {
		// 目标为组播地址，广播所有的通信中转服务
		return d.multicastBCAddr, nil

	} else {
		return "", ErrIncorrectDestAddress
	}
}

func (d *_ForwardOutDeliverer) getDistEntityForwardAddr(entId uid.Id) (string, error) {
	dent, ok := d.dentq.GetDistEntity(entId)
	if !ok {
		return "", ErrDistEntityNotFound
	}

	idx := pie.FindFirstUsing(dent.Nodes, func(node dentq.Node) bool {
		return node.Service == d.gate
	})
	if idx < 0 {
		return "", ErrDistEntityNodeNotFound
	}

	return dent.Nodes[idx].RemoteAddr, nil
}
