package processor

import (
	"errors"
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

var (
	ErrDistEntityNotFound       = errors.New("rpc: distributed entity node not found")
	ErrForwardToServiceNotFound = errors.New("rpc: the forwarding to the service node is not found")
	ErrIncorrectDestAddress     = errors.New("rpc: incorrect destination AddressDetails")
)

// NewForwardingDeliverer RPC转发器，用于S->C的通信
func NewForwardingDeliverer(forwardService string) IDeliverer {
	return &_ForwardingDeliverer{forwardService: forwardService}
}

// _ForwardingDeliverer RPC转发器，用于S->C的通信
type _ForwardingDeliverer struct {
	servCtx         service.Context
	dist            dserv.IDistService
	dentq           dentq.IDistEntityQuerier
	forwardService  string
	multicastBCAddr string
}

// Init 初始化
func (d *_ForwardingDeliverer) Init(ctx service.Context) {
	d.servCtx = ctx
	d.dist = dserv.Using(ctx)
	d.dentq = dentq.Using(ctx)
	d.multicastBCAddr = d.dist.MakeBroadcastAddr(d.forwardService)

	log.Debugf(d.servCtx, "rpc deliverer %q started", types.AnyFullName(*d))
}

// Shut 结束
func (d *_ForwardingDeliverer) Shut(ctx service.Context) {
	log.Debugf(d.servCtx, "rpc deliverer %q stopped", types.AnyFullName(*d))
}

// Match 是否匹配
func (d *_ForwardingDeliverer) Match(ctx service.Context, dst, path string, oneWay bool) bool {
	if !gate.ClientAddressDetails.InDomain(dst) {
		return false
	}

	if !oneWay {
		if !gate.ClientAddressDetails.InNodeSubdomain(dst) {
			return false
		}
	}

	return true
}

// Request 请求
func (d *_ForwardingDeliverer) Request(ctx service.Context, dst, path string, args []any) runtime.AsyncRet {
	ret := concurrent.MakeRespAsyncRet()
	future := concurrent.MakeFuture(d.dist.GetFutures(), nil, ret)

	forwardAddr, err := d.getForwardAddr(uid.From(netpath.Base(gate.ClientAddressDetails.PathSeparator, dst)))
	if err != nil {
		future.Cancel(err)
		return ret.CastAsyncRet()
	}

	vargs, err := variant.MakeArray(args)
	if err != nil {
		future.Cancel(err)
		return ret.CastAsyncRet()
	}

	raw := &gap.MsgRPCRequest{
		CorrId: future.Id,
		Path:   path,
		Args:   vargs,
	}

	buf := binaryutil.MakeRecycleBytes(binaryutil.BytesPool.Get(raw.Size()))
	defer buf.Release()

	msg := &gap.MsgForward{
		Dst:     dst,
		CorrId:  raw.CorrId,
		RawId:   gap.MsgId_RPC_Request,
		RawData: buf.Data(),
	}

	if err = d.dist.SendMsg(forwardAddr, msg); err != nil {
		future.Cancel(err)
		return ret.CastAsyncRet()
	}

	log.Debugf(d.servCtx, "rpc request(%d) forwarding to dst:%q, path:%q ok", future.Id, forwardAddr, path)

	return ret.CastAsyncRet()
}

// Notify 通知
func (d *_ForwardingDeliverer) Notify(ctx service.Context, dst, path string, args []any) error {
	forwardAddr, err := func() (string, error) {
		if gate.ClientAddressDetails.InNodeSubdomain(dst) {
			// 目标为单播地址，查询实体的通信中转服务地址
			return d.getForwardAddr(uid.From(netpath.Base(gate.ClientAddressDetails.PathSeparator, dst)))
		} else if gate.ClientAddressDetails.InMulticastSubdomain(dst) {
			// 目标为组播地址，广播所有的通信中转服务
			return d.multicastBCAddr, nil
		} else {
			return "", ErrIncorrectDestAddress
		}
	}()
	if err != nil {
		return err
	}

	vargs, err := variant.MakeArray(args)
	if err != nil {
		return err
	}

	raw := &gap.MsgOneWayRPC{
		Path: path,
		Args: vargs,
	}

	buf := binaryutil.MakeRecycleBytes(binaryutil.BytesPool.Get(raw.Size()))
	defer buf.Release()

	msg := &gap.MsgForward{
		Dst:     dst,
		RawId:   gap.MsgId_OneWayRPC,
		RawData: buf.Data(),
	}

	if err = d.dist.SendMsg(forwardAddr, msg); err != nil {
		return err
	}

	log.Debugf(d.servCtx, "rpc notify forwarding to dst:%q, path:%q ok", forwardAddr, path)

	return nil
}

func (d *_ForwardingDeliverer) getForwardAddr(entId uid.Id) (string, error) {
	dent, ok := d.dentq.GetDistEntity(entId)
	if !ok {
		return "", ErrDistEntityNotFound
	}

	idx := pie.FindFirstUsing(dent.Nodes, func(node dentq.Node) bool {
		return node.Service == d.forwardService
	})
	if idx < 0 {
		return "", ErrForwardToServiceNotFound
	}

	return dent.Nodes[idx].RemoteAddr, nil
}
