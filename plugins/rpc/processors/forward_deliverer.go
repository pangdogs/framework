package processors

import (
	"errors"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/types"
	"git.golaxy.org/core/util/uid"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/variant"
	"git.golaxy.org/framework/plugins/dentq"
	"git.golaxy.org/framework/plugins/dserv"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/util/binaryutil"
	"git.golaxy.org/framework/util/concurrent"
	"git.golaxy.org/framework/util/pathutil"
	"github.com/elliotchance/pie/v2"
)

var (
	ErrDistEntityNotFound       = errors.New("rpc: distributed entity node not found")
	ErrForwardToServiceNotFound = errors.New("rpc: the forwarding to the service node is not found")
	ErrIncorrectDestAddress     = errors.New("rpc: incorrect destination Address")
)

// ForwardDeliverer 转发RPC的投递器
type ForwardDeliverer struct {
	AcceptDomain             string // 需要转发的主域
	AcceptNodeSubdomain      string // 需要转发的节点地址子域
	AcceptMulticastSubdomain string // 需要转发的组播地址子域
	AcceptPathSeparator      string // 需要转发的地址路径分隔符
	ForwardTo                string // 转发的目标服务
	servCtx                  service.Context
	dist                     dserv.IDistService
	dentq                    dentq.IDistEntityQuerier
	forwardBCAddr            string
}

// Init 初始化
func (d *ForwardDeliverer) Init(ctx service.Context) {
	d.servCtx = ctx
	d.dist = dserv.Using(ctx)
	d.dentq = dentq.Using(ctx)

	d.forwardBCAddr = d.dist.MakeBroadcastAddr(d.ForwardTo)

	log.Debugf(d.servCtx, "rpc deliverer %q started", types.AnyFullName(*d))
}

// Shut 结束
func (d *ForwardDeliverer) Shut(ctx service.Context) {
	log.Debugf(d.servCtx, "rpc deliverer %q stopped", types.AnyFullName(*d))
}

// Match 是否匹配
func (d *ForwardDeliverer) Match(ctx service.Context, dst, path string, oneWay bool) bool {
	if !pathutil.InDir(d.AcceptPathSeparator, dst, d.AcceptDomain) {
		return false
	}

	if !oneWay {
		if !pathutil.InDir(d.AcceptPathSeparator, dst, d.AcceptNodeSubdomain) {
			return false
		}
	}

	return true
}

// Request 请求
func (d *ForwardDeliverer) Request(ctx service.Context, dst, path string, args []any) runtime.AsyncRet {
	ret := concurrent.MakeRespAsyncRet()
	future := concurrent.MakeFuture(d.dist.GetFutures(), nil, ret)

	forwardAddr, err := d.getForwardAddr(uid.From(pathutil.Base(d.AcceptPathSeparator, dst)))
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
		Dst: dst,
		Raw: buf.Data(),
	}

	if err = d.dist.SendMsg(forwardAddr, msg); err != nil {
		future.Cancel(err)
		return ret.CastAsyncRet()
	}

	log.Debugf(d.servCtx, "rpc request(%d) forwarding to dst:%q, path:%q ok", future.Id, forwardAddr, path)

	return ret.CastAsyncRet()
}

// Notify 通知
func (d *ForwardDeliverer) Notify(ctx service.Context, dst, path string, args []any) error {
	forwardAddr, err := func() (string, error) {
		if pathutil.InDir(d.AcceptPathSeparator, dst, d.AcceptNodeSubdomain) {
			return d.getForwardAddr(uid.From(pathutil.Base(d.AcceptPathSeparator, dst)))
		} else if pathutil.InDir(d.AcceptPathSeparator, dst, d.AcceptMulticastSubdomain) {
			return d.forwardBCAddr, nil
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
		Dst: dst,
		Raw: buf.Data(),
	}

	if err = d.dist.SendMsg(forwardAddr, msg); err != nil {
		return err
	}

	log.Debugf(d.servCtx, "rpc notify forwarding to dst:%q, path:%q ok", forwardAddr, path)

	return nil
}

func (d *ForwardDeliverer) getForwardAddr(entId uid.Id) (string, error) {
	dent, ok := d.dentq.GetDistEntity(entId)
	if !ok {
		return "", ErrDistEntityNotFound
	}

	idx := pie.FindFirstUsing(dent.Nodes, func(node dentq.Node) bool {
		return node.Service == d.ForwardTo
	})
	if idx < 0 {
		return "", ErrForwardToServiceNotFound
	}

	return dent.Nodes[idx].RemoteAddr, nil
}
