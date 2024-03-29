package processor

import (
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/generic"
	"git.golaxy.org/core/util/types"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/codec"
	"git.golaxy.org/framework/net/gap/variant"
	"git.golaxy.org/framework/plugins/dentq"
	"git.golaxy.org/framework/plugins/dserv"
	"git.golaxy.org/framework/plugins/gate"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/plugins/router"
	"github.com/elliotchance/pie/v2"
)

// NewInboundDispatcher 创建入站方向RPC分发器，用于C->S的通信
func NewInboundDispatcher(mc gap.IMsgCreator) IDispatcher {
	return &_InboundDispatcher{
		encoder: codec.MakeEncoder(),
		decoder: codec.MakeDecoder(mc),
	}
}

// _InboundDispatcher 入站方向RPC分发器，用于C->S的通信
type _InboundDispatcher struct {
	servCtx service.Context
	dist    dserv.IDistService
	gate    gate.IGate
	router  router.IRouter
	dentq   dentq.IDistEntityQuerier
	watcher gate.IWatcher
	encoder codec.Encoder
	decoder codec.Decoder
}

// Init 初始化
func (d *_InboundDispatcher) Init(ctx service.Context) {
	d.servCtx = ctx
	d.dist = dserv.Using(ctx)
	d.gate = gate.Using(ctx)
	d.router = router.Using(ctx)
	d.dentq = dentq.Using(ctx)
	d.watcher = d.gate.Watch(ctx, generic.CastDelegateAction3(d.handleSessionChanged))

	log.Debugf(d.servCtx, "rpc dispatcher %q started", types.AnyFullName(*d))
}

// Shut 结束
func (d *_InboundDispatcher) Shut(ctx service.Context) {
	<-d.watcher.Terminate()

	log.Debugf(d.servCtx, "rpc dispatcher %q stopped", types.AnyFullName(*d))
}

func (d *_InboundDispatcher) handleSessionChanged(session gate.ISession, curState gate.SessionState, lastState gate.SessionState) {
	switch curState {
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
		return d.acceptForward(session, mp.Head.Seq, mp.Msg.(*gap.MsgForward))
	}

	return nil
}

func (d *_InboundDispatcher) acceptForward(session gate.ISession, seq int64, req *gap.MsgForward) error {
	switch req.TransId {
	case gap.MsgId_RPC_Request, gap.MsgId_RPC_Reply, gap.MsgId_OneWayRPC:
		break
	default:
		return nil
	}

	entity, cliAddr, ok := d.router.LookupEntity(session.GetId())
	if !ok {
		go d.forwardingFinish(session, req.Dst, req.CorrId, ErrEntityNotFound)
		return ErrEntityNotFound
	}

	distEntity, ok := d.dentq.GetDistEntity(entity.GetId())
	if !ok {
		go d.forwardingFinish(session, req.Dst, req.CorrId, ErrDistEntityNotFound)
		return ErrDistEntityNotFound
	}

	nodeIdx := pie.FindFirstUsing(distEntity.Nodes, func(node dentq.Node) bool {
		return node.Service == req.Dst
	})
	if nodeIdx < 0 {
		go d.forwardingFinish(session, req.Dst, req.CorrId, ErrDistEntityNodeNotFound)
		return ErrDistEntityNodeNotFound
	}
	node := distEntity.Nodes[nodeIdx]

	msg := &gap.MsgForward{
		Transit:   d.dist.GetNodeDetails().LocalAddr, // 中转地址
		Dst:       entity.GetId().String(),           // 目标实体
		TransId:   req.TransId,
		TransData: req.TransData,
	}

	if err := d.dist.ForwardMsg(cliAddr, node.RemoteAddr, seq, msg); err != nil {
		go d.forwardingFinish(session, node.RemoteAddr, req.CorrId, err)
		return err
	}

	go d.forwardingFinish(session, req.Dst, req.CorrId, nil)
	return nil
}

func (d *_InboundDispatcher) forwardingFinish(session gate.ISession, dst string, corrId int64, err error) {
	if err == nil {
		if corrId != 0 {
			log.Debugf(d.servCtx, "forwarding session:%q rpc request(%d) to dst:%q finish", session.GetId(), corrId, dst)
		} else {
			log.Debugf(d.servCtx, "forwarding session:%q rpc notify to dst:%q finish", session.GetId(), dst)
		}
	} else {
		if corrId != 0 {
			log.Errorf(d.servCtx, "forwarding session:%q rpc request(%d) to dst:%q failed, %s", session.GetId(), corrId, dst, err)
			d.reply(session, corrId, err)
		} else {
			log.Errorf(d.servCtx, "forwarding session:%q rpc notify to dst:%q failed, %s", session.GetId(), dst, err)
		}
	}
}

func (d *_InboundDispatcher) reply(session gate.ISession, corrId int64, retErr error) {
	if corrId == 0 || retErr == nil {
		return
	}

	msg := &gap.MsgRPCReply{
		CorrId: corrId,
		Error:  *variant.MakeError(retErr),
	}

	bs, err := d.encoder.EncodeBytes(d.dist.GetNodeDetails().LocalAddr, 0, msg)
	if err != nil {
		log.Errorf(d.servCtx, "rpc reply(%d) to session:%q failed, %s", corrId, session.GetId(), err)
		return
	}
	defer bs.Release()

	err = session.SendData(bs.Data())
	if err != nil {
		log.Errorf(d.servCtx, "rpc reply(%d) to session:%q failed, %s", corrId, session.GetId(), err)
		return
	}

	log.Debugf(d.servCtx, "rpc reply(%d) to session:%q ok", corrId, session.GetId())
}
