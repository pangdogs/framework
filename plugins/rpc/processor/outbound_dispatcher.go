package processor

import (
	"context"
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/generic"
	"git.golaxy.org/core/util/types"
	"git.golaxy.org/core/util/uid"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/codec"
	"git.golaxy.org/framework/net/gap/variant"
	"git.golaxy.org/framework/net/netpath"
	"git.golaxy.org/framework/plugins/dserv"
	"git.golaxy.org/framework/plugins/gate"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/plugins/router"
)

// NewOutboundDispatcher 创建出站方向RPC分发器，用于S->C的通信
func NewOutboundDispatcher() IDispatcher {
	return &_OutboundDispatcher{
		encoder: codec.MakeEncoder(),
	}
}

// _OutboundDispatcher 出站方向RPC分发器，用于S->C的通信
type _OutboundDispatcher struct {
	servCtx service.Context
	dist    dserv.IDistService
	router  router.IRouter
	encoder codec.Encoder
	watcher dserv.IWatcher
}

// Init 初始化
func (d *_OutboundDispatcher) Init(ctx service.Context) {
	d.servCtx = ctx
	d.dist = dserv.Using(ctx)
	d.router = router.Using(ctx)
	d.watcher = d.dist.WatchMsg(context.Background(), generic.CastDelegateFunc2(d.handleMsg))

	log.Debugf(d.servCtx, "rpc dispatcher %q started", types.AnyFullName(*d))
}

// Shut 结束
func (d *_OutboundDispatcher) Shut(ctx service.Context) {
	<-d.watcher.Terminate()

	log.Debugf(d.servCtx, "rpc dispatcher %q stopped", types.AnyFullName(*d))
}

func (d *_OutboundDispatcher) handleMsg(topic string, mp gap.MsgPacket) error {
	// 只支持客户端域通信
	if !d.dist.GetAddressDetails().InDomain(mp.Head.Src) {
		return nil
	}

	switch mp.Head.MsgId {
	case gap.MsgId_Forward:
		d.acceptForward(mp.Head.Src, mp.Msg.(*gap.MsgForward))
	}

	return nil
}

func (d *_OutboundDispatcher) acceptForward(src string, req *gap.MsgForward) {
	if gate.ClientAddressDetails.InNodeSubdomain(req.Dst) {
		// 目标为单播地址，解析实体Id
		entId := uid.From(netpath.Base(gate.ClientAddressDetails.PathSeparator, req.Dst))

		// 为了保持消息时序，在实体线程中，向对端发送消息
		asyncRet := d.servCtx.Call(entId, func(entity ec.Entity, _ ...any) runtime.Ret {
			session, ok := d.router.LookupSession(entity.GetId())
			if !ok {
				return runtime.MakeRet(nil, ErrSessionNotFound)
			}

			bs, err := d.encoder.EncodeBytes(src, 0, &gap.MsgRaw{Id: req.TransId, Data: req.TransData})
			if err != nil {
				return runtime.MakeRet(nil, err)
			}
			defer bs.Release()

			err = session.SendData(bs.Data())
			if err != nil {
				return runtime.MakeRet(nil, err)
			}

			return runtime.MakeRet(nil, nil)
		})
		go d.forwardingFinish(src, req, (<-asyncRet).Error)
		return

	} else if gate.ClientAddressDetails.InMulticastSubdomain(req.Dst) {
		// 目标为组播地址，解析分组Id
		groupId := uid.From(netpath.Base(gate.ClientAddressDetails.PathSeparator, req.Dst))

		group, ok := d.router.GetGroup(groupId)
		if !ok {
			go d.forwardingFinish(src, req, ErrGroupNotFound)
			return
		}

		bs, err := d.encoder.EncodeBytes(src, 0, &gap.MsgRaw{Id: req.TransId, Data: req.TransData})
		if err != nil {
			go d.forwardingFinish(src, req, err)
			return
		}

		// 为了保持消息时序，使用分组发送数据的channel
		select {
		case group.SendDataChan() <- bs:
			go d.forwardingFinish(src, req, nil)
		default:
			bs.Release()
			go d.forwardingFinish(src, req, ErrGroupChanIsFull)
		}
		return

	} else {
		go d.forwardingFinish(src, req, ErrIncorrectDestAddress)
		return
	}
}

func (d *_OutboundDispatcher) forwardingFinish(src string, req *gap.MsgForward, err error) {
	if err == nil {
		if req.CorrId != 0 {
			log.Debugf(d.servCtx, "forwarding src:%q rpc request(%d) to remote:%q finish", src, req.CorrId, req.Dst)
		} else {
			log.Debugf(d.servCtx, "forwarding src:%q rpc notify to remote:%q finish", src, req.Dst)
		}
	} else {
		if req.CorrId != 0 {
			log.Errorf(d.servCtx, "forwarding src:%q rpc request(%d) to remote:%q failed, %s", src, req.CorrId, req.Dst, err)
			d.reply(src, req.CorrId, err)
		} else {
			log.Errorf(d.servCtx, "forwarding src:%q rpc notify to remote:%q failed, %s", src, req.Dst, err)
		}
	}
}

func (d *_OutboundDispatcher) reply(src string, corrId int64, retErr error) {
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
