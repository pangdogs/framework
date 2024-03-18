package processor

import (
	"errors"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/generic"
	"git.golaxy.org/core/util/types"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/codec"
	"git.golaxy.org/framework/net/gap/variant"
	"git.golaxy.org/framework/plugins/dserv"
	"git.golaxy.org/framework/plugins/gate"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/plugins/router"
)

var (
	ErrEntityNotFound = errors.New("rpc: session routing to entity not found")
)

// NewInboundDispatcher 创建入站方向RPC分发器，用于C->S的通信
func NewInboundDispatcher(mc gap.IMsgCreator) IDispatcher {
	return &_InboundDispatcher{
		decoder: codec.MakeDecoder(mc),
	}
}

// _InboundDispatcher 入站方向RPC分发器，用于C->S的通信
type _InboundDispatcher struct {
	servCtx service.Context
	dist    dserv.IDistService
	gate    gate.IGate
	router  router.IRouter
	watcher gate.IWatcher
	decoder codec.Decoder
}

// Init 初始化
func (d *_InboundDispatcher) Init(ctx service.Context) {
	d.servCtx = ctx
	d.dist = dserv.Using(ctx)
	d.gate = gate.Using(ctx)
	d.router = router.Using(ctx)
	d.watcher = d.gate.Watch(ctx, generic.CastDelegateAction3(d.handleSessionChanged))

	log.Debugf(d.servCtx, "rpc dispatcher %q started", types.AnyFullName(*d))
}

// Shut 结束
func (d *_InboundDispatcher) Shut(ctx service.Context) {
	<-d.watcher.Terminate()

	log.Debugf(d.servCtx, "rpc dispatcher %q stopped", types.AnyFullName(*d))
}

func (d *_InboundDispatcher) handleSessionChanged(session gate.ISession, newState gate.SessionState, oldState gate.SessionState) {
	switch newState {
	case gate.SessionState_Confirmed:
		err := session.GetSettings().
			RecvDataHandler(append(session.GetSettings().CurrRecvDataHandler, d.handleRecvData)).
			Change()
		if err != nil {
			log.Errorf(d.servCtx, "change session %q settings failed, %s", session.GetId(), err)
			return
		}
	}
}

func (d *_InboundDispatcher) handleRecvData(session gate.ISession, data []byte) error {
	mp, err := d.decoder.DecodeBytes(data)
	if err != nil {
		return err
	}

	switch mp.Head.MsgId {
	case gap.MsgId_Forward:
		return d.acceptForward(session, mp.Head.Src, mp.Head.Seq, mp.Msg.(*gap.MsgForward))
	}

	return nil
}

func (d *_InboundDispatcher) acceptForward(session gate.ISession, src string, seq int64, req *gap.MsgForward) error {
	_, ok := d.router.LookupEntity(session.GetId())
	if !ok {
		go d.forwardingFinish(src, req.Dst, req.CorrId, ErrEntityNotFound)
		return ErrEntityNotFound
	}

	addr := d.dist.GetAddressDetails()

	if addr.InBroadcastSubdomain(req.Dst) || addr.InBalanceSubdomain(req.Dst) || addr.InNodeSubdomain(req.Dst) {
		msg := &gap.MsgRaw{
			Id:   req.RawId,
			Data: req.RawData,
		}

		if err := d.dist.ForwardMsg(src, req.Dst, seq, msg); err != nil {
			go d.forwardingFinish(src, req.Dst, req.CorrId, err)
			return err
		}

		return nil

	} else {
		go d.forwardingFinish(src, req.Dst, req.CorrId, ErrIncorrectDestAddress)
		return ErrIncorrectDestAddress
	}
}

func (d *_InboundDispatcher) forwardingFinish(src, dst string, corrId int64, err error) {
	if err == nil {
		if corrId != 0 {
			log.Debugf(d.servCtx, "forwarding src:%q rpc request(%d) to remote:%q finish", src, corrId, dst)
		} else {
			log.Debugf(d.servCtx, "forwarding src:%q rpc notify to remote:%q finish", src, dst)
		}
	} else {
		if corrId != 0 {
			log.Errorf(d.servCtx, "forwarding src:%q rpc request(%d) to remote:%q failed, %s", src, corrId, dst, err)
			d.reply(src, corrId, err)
		} else {
			log.Errorf(d.servCtx, "forwarding src:%q rpc notify to remote:%q failed, %s", src, dst, err)
		}
	}
}

func (d *_InboundDispatcher) reply(src string, corrId int64, retErr error) {
	if corrId == 0 || retErr == nil {
		return
	}

	msg := &gap.MsgRPCReply{
		CorrId: corrId,
		Error:  *variant.MakeError(retErr),
	}

	err := d.dist.SendMsg(src, msg)
	if err != nil {
		log.Errorf(d.servCtx, "rpc reply(%d) to src:%q failed, %s", corrId, src, err)
		return
	}

	log.Debugf(d.servCtx, "rpc reply(%d) to src:%q ok", corrId, src)
}
