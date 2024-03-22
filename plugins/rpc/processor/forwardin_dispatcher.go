package processor

import (
	"context"
	"fmt"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/generic"
	"git.golaxy.org/core/util/types"
	"git.golaxy.org/core/util/uid"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/codec"
	"git.golaxy.org/framework/net/gap/variant"
	"git.golaxy.org/framework/plugins/dserv"
	"git.golaxy.org/framework/plugins/gate"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/plugins/rpc/callpath"
	"git.golaxy.org/framework/util/binaryutil"
	"git.golaxy.org/framework/util/concurrent"
	"reflect"
)

// NewForwardInDispatcher RPC内转分发器，用于C->S的通信
func NewForwardInDispatcher(mc gap.IMsgCreator) IDispatcher {
	return &_ForwardInDispatcher{
		encoder: codec.MakeEncoder(),
		decoder: codec.MakeDecoder(mc),
	}
}

// _ForwardInDispatcher RPC内转分发器，用于C->S的通信
type _ForwardInDispatcher struct {
	servCtx service.Context
	dist    dserv.IDistService
	watcher dserv.IWatcher
	encoder codec.Encoder
	decoder codec.Decoder
}

// Init 初始化
func (d *_ForwardInDispatcher) Init(ctx service.Context) {
	d.servCtx = ctx
	d.dist = dserv.Using(ctx)
	d.watcher = d.dist.WatchMsg(context.Background(), generic.CastDelegateFunc2(d.handleMsg))

	log.Debugf(d.servCtx, "rpc dispatcher %q started", types.AnyFullName(*d))
}

// Shut 结束
func (d *_ForwardInDispatcher) Shut(ctx service.Context) {
	<-d.watcher.Terminate()

	log.Debugf(d.servCtx, "rpc dispatcher %q stopped", types.AnyFullName(*d))
}

func (d *_ForwardInDispatcher) handleMsg(topic string, mp gap.MsgPacket) error {
	// 只支持客户端域通信
	if !gate.ClientAddressDetails.InDomain(mp.Head.Src) {
		return nil
	}

	switch mp.Head.MsgId {
	case gap.MsgId_Forward:
		return d.acceptForward(mp.Head.Src, mp.Msg.(*gap.MsgForward))
	}

	return nil
}

func (d *_ForwardInDispatcher) acceptForward(src string, req *gap.MsgForward) error {
	switch req.TransId {
	case gap.MsgId_OneWayRPC:
		msg := &gap.MsgOneWayRPC{}

		if _, err := msg.Write(req.TransData); err != nil {
			return fmt.Errorf("unmarshal msg(%d) failed, %s", msg.MsgId(), err)
		}

		return d.acceptNotify(msg)

	case gap.MsgId_RPC_Request:
		msg := &gap.MsgRPCRequest{}

		if _, err := msg.Write(req.TransData); err != nil {
			go d.reply(src, req.Gate, req.CorrId, nil, err)
			return fmt.Errorf("unmarshal msg(%d) failed, %s", msg.MsgId(), err)
		}

		return d.acceptRequest(src, req.Gate, msg)

	case gap.MsgId_RPC_Reply:
		msg := &gap.MsgRPCReply{}

		if _, err := msg.Write(req.TransData); err != nil {
			return fmt.Errorf("unmarshal msg(%d) failed, %s", msg.MsgId(), err)
		}

		return d.resolve(msg)
	}

	return nil
}

func (d *_ForwardInDispatcher) acceptNotify(req *gap.MsgOneWayRPC) error {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		return fmt.Errorf("parse rpc notify path:%q failed, %s", req.Path, err)
	}

	switch cp.Category {
	case callpath.Service:
		go func() {
			if _, err := CallService(d.servCtx, cp.Plugin, cp.Method, req.Args); err != nil {
				log.Errorf(d.servCtx, "rpc notify service plugin:%q, method:%q calls failed, %s", cp.Plugin, cp.Method, err)
				return
			}
			log.Debugf(d.servCtx, "rpc notify service plugin:%q, method:%q calls finished", cp.Plugin, cp.Method)
			return
		}()

		return nil

	case callpath.Runtime:
		asyncRet, err := CallRuntime(d.servCtx, uid.From(cp.EntityId), cp.Plugin, cp.Method, req.Args)
		if err != nil {
			log.Errorf(d.servCtx, "rpc notify entity:%q, runtime plugin:%q, method:%q calls failed, %s", cp.EntityId, cp.Plugin, cp.Method, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(d.servCtx)
			if !ret.OK() {
				log.Errorf(d.servCtx, "rpc notify entity:%q, runtime plugin:%q, method:%q calls failed, %s", cp.EntityId, cp.Plugin, cp.Method, ret.Error)
				return
			}
			log.Debugf(d.servCtx, "rpc notify entity:%q, runtime plugin:%q, method:%q calls finished", cp.EntityId, cp.Plugin, cp.Method)
		}()

		return nil

	case callpath.Entity:
		asyncRet, err := CallEntity(d.servCtx, uid.From(cp.EntityId), cp.Component, cp.Method, req.Args)
		if err != nil {
			log.Errorf(d.servCtx, "rpc notify entity:%q, component:%q, method:%q calls failed, %s", cp.EntityId, cp.Component, cp.Method, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(d.servCtx)
			if !ret.OK() {
				log.Errorf(d.servCtx, "rpc notify entity:%q, component:%q, method:%q calls failed, %s", cp.EntityId, cp.Component, cp.Method, ret.Error)
				return
			}
			log.Debugf(d.servCtx, "rpc notify entity:%q, component:%q, method:%q calls finished", cp.EntityId, cp.Component, cp.Method)
		}()

		return nil
	}

	return nil
}

func (d *_ForwardInDispatcher) acceptRequest(src, gate string, req *gap.MsgRPCRequest) error {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		err = fmt.Errorf("parse rpc request(%d) path %q failed, %s", req.CorrId, req.Path, err)
		go d.reply(src, gate, req.CorrId, nil, err)
		return err
	}

	switch cp.Category {
	case callpath.Service:
		go func() {
			retsRV, err := CallService(d.servCtx, cp.Plugin, cp.Method, req.Args)
			if err != nil {
				log.Errorf(d.servCtx, "rpc request(%d) service plugin:%q, method:%q calls failed, %s", req.CorrId, cp.Plugin, cp.Method, err)
				d.reply(src, gate, req.CorrId, nil, err)
				return
			}
			log.Debugf(d.servCtx, "rpc request(%d) service plugin:%q, method:%q calls finished", req.CorrId, cp.Plugin, cp.Method)
			d.reply(src, gate, req.CorrId, retsRV, nil)
		}()

		return nil

	case callpath.Runtime:
		asyncRet, err := CallRuntime(d.servCtx, uid.From(cp.EntityId), cp.Plugin, cp.Method, req.Args)
		if err != nil {
			log.Errorf(d.servCtx, "rpc request(%d) entity:%q, runtime plugin:%q, method:%q calls failed, %s", req.CorrId, cp.EntityId, cp.Plugin, cp.Method, err)
			go d.reply(src, gate, req.CorrId, nil, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(d.servCtx)
			if !ret.OK() {
				log.Errorf(d.servCtx, "rpc request(%d) entity:%q, runtime plugin:%q, method:%q calls failed, %s", req.CorrId, cp.EntityId, cp.Plugin, cp.Method, ret.Error)
				d.reply(src, gate, req.CorrId, nil, ret.Error)
				return
			}
			log.Debugf(d.servCtx, "rpc request(%d) entity:%q, runtime plugin:%q, method:%q calls finished", req.CorrId, cp.EntityId, cp.Plugin, cp.Method)
			d.reply(src, gate, req.CorrId, ret.Value.([]reflect.Value), nil)
		}()

		return nil

	case callpath.Entity:
		asyncRet, err := CallEntity(d.servCtx, uid.From(cp.EntityId), cp.Component, cp.Method, req.Args)
		if err != nil {
			log.Errorf(d.servCtx, "rpc request(%d) entity:%q, component:%q, method:%q calls failed, %s", req.CorrId, cp.EntityId, cp.Component, cp.Method, err)
			go d.reply(src, gate, req.CorrId, nil, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(d.servCtx)
			if !ret.OK() {
				log.Errorf(d.servCtx, "rpc request(%d) entity:%q, component:%q, method:%q calls failed, %s", req.CorrId, cp.EntityId, cp.Component, cp.Method, ret.Error)
				d.reply(src, gate, req.CorrId, nil, ret.Error)
				return
			}
			log.Debugf(d.servCtx, "rpc request(%d) entity:%q, component:%q, method:%q calls finished", req.CorrId, cp.EntityId, cp.Component, cp.Method)
			d.reply(src, gate, req.CorrId, ret.Value.([]reflect.Value), nil)
		}()

		return nil
	}

	return nil
}

func (d *_ForwardInDispatcher) reply(src, gate string, corrId int64, retsRV []reflect.Value, retErr error) {
	if corrId == 0 {
		return
	}

	msg := &gap.MsgRPCReply{
		CorrId: corrId,
	}

	var err error
	msg.Rets, err = variant.MakeArray(retsRV)
	if err != nil {
		log.Errorf(d.servCtx, "rpc reply(%d) to src:%q failed, %s", corrId, src, err)
		return
	}

	if retErr != nil {
		msg.Error = *variant.MakeError(retErr)
	}

	bs := binaryutil.MakeRecycleBytes(binaryutil.BytesPool.Get(msg.Size()))
	defer bs.Release()

	if _, err = msg.Read(bs.Data()); err != nil {
		log.Errorf(d.servCtx, "rpc reply(%d) to src:%q failed, %s", corrId, src, err)
		return
	}

	forwardMsg := &gap.MsgForward{
		Dst:       src,
		CorrId:    corrId,
		TransId:   msg.MsgId(),
		TransData: bs.Data(),
	}

	err = d.dist.SendMsg(gate, forwardMsg)
	if err != nil {
		log.Errorf(d.servCtx, "rpc reply(%d) to src:%q failed, %s", corrId, src, err)
		return
	}

	log.Debugf(d.servCtx, "rpc reply(%d) to src:%q ok", corrId, src)
}

func (d *_ForwardInDispatcher) resolve(reply *gap.MsgRPCReply) error {
	ret := concurrent.Ret[any]{}

	if reply.Error.OK() {
		if len(reply.Rets) > 0 {
			ret.Value = reply.Rets
		}
	} else {
		ret.Error = &reply.Error
	}

	return d.dist.GetFutures().Resolve(reply.CorrId, ret)
}
