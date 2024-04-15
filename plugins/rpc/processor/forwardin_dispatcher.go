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
	d.watcher = d.dist.WatchMsg(context.Background(), generic.MakeDelegateFunc2(d.handleMsg))

	log.Debugf(d.servCtx, "rpc dispatcher %q started", types.AnyFullName(*d))
}

// Shut 结束
func (d *_ForwardInDispatcher) Shut(ctx service.Context) {
	<-d.watcher.Terminate()

	log.Debugf(d.servCtx, "rpc dispatcher %q stopped", types.AnyFullName(*d))
}

func (d *_ForwardInDispatcher) handleMsg(topic string, mp gap.MsgPacket) error {
	// 只支持客户端域通信
	if !gate.CliDetails.InDomain(mp.Head.Src) {
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
		if err := gap.Unmarshal(msg, req.TransData); err != nil {
			return err
		}
		return d.acceptNotify(src, req.Dst, req.Transit, msg)

	case gap.MsgId_RPC_Request:
		msg := &gap.MsgRPCRequest{}
		if err := gap.Unmarshal(msg, req.TransData); err != nil {
			go d.reply(src, req.Transit, req.CorrId, nil, err)
			return err
		}
		return d.acceptRequest(src, req.Dst, req.Transit, msg)

	case gap.MsgId_RPC_Reply:
		msg := &gap.MsgRPCReply{}
		if err := gap.Unmarshal(msg, req.TransData); err != nil {
			return err
		}
		return d.resolve(msg)
	}

	return nil
}

func (d *_ForwardInDispatcher) acceptNotify(src, dst, transit string, req *gap.MsgOneWayRPC) error {
	cp, err := d.parseCallPath(dst, req.Path)
	if err != nil {
		return fmt.Errorf("parse rpc notify failed, src:%q, dst:%q, transit:%q, path:%q, %s", src, dst, transit, req.Path, err)
	}

	switch cp.Category {
	case callpath.Service:
		go func() {
			if _, err := CallService(d.servCtx, cp.Plugin, cp.Method, req.Args); err != nil {
				log.Errorf(d.servCtx, "rpc notify service plugin:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", cp.Plugin, cp.Method, src, dst, transit, req.Path, err)
			} else {
				log.Debugf(d.servCtx, "rpc notify service plugin:%q, method:%q calls finished, src:%q, dst:%q, transit:%q, path:%q", cp.Plugin, cp.Method, src, dst, transit, req.Path)
			}
		}()

		return nil

	case callpath.Runtime:
		asyncRet, err := CallRuntime(d.servCtx, cp.EntityId, cp.Plugin, cp.Method, req.Args)
		if err != nil {
			log.Errorf(d.servCtx, "rpc notify entity:%q, runtime plugin:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", cp.EntityId, cp.Plugin, cp.Method, src, dst, transit, req.Path, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(d.servCtx)
			if !ret.OK() {
				log.Errorf(d.servCtx, "rpc notify entity:%q, runtime plugin:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", cp.EntityId, cp.Plugin, cp.Method, src, dst, transit, req.Path, ret.Error)
			} else {
				log.Debugf(d.servCtx, "rpc notify entity:%q, runtime plugin:%q, method:%q calls finished, src:%q, dst:%q, transit:%q, path:%q,", cp.EntityId, cp.Plugin, cp.Method, src, dst, transit, req.Path)
			}
		}()

		return nil

	case callpath.Entity:
		asyncRet, err := CallEntity(d.servCtx, cp.EntityId, cp.Component, cp.Method, req.Args)
		if err != nil {
			log.Errorf(d.servCtx, "rpc notify entity:%q, component:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", cp.EntityId, cp.Component, cp.Method, src, dst, transit, req.Path, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(d.servCtx)
			if !ret.OK() {
				log.Errorf(d.servCtx, "rpc notify entity:%q, component:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", cp.EntityId, cp.Component, cp.Method, src, dst, transit, req.Path, ret.Error)
			} else {
				log.Debugf(d.servCtx, "rpc notify entity:%q, component:%q, method:%q calls finished, src:%q, dst:%q, transit:%q, path:%q", cp.EntityId, cp.Component, cp.Method, src, dst, transit, req.Path)
			}
		}()

		return nil
	}

	return nil
}

func (d *_ForwardInDispatcher) acceptRequest(src, dst, transit string, req *gap.MsgRPCRequest) error {
	cp, err := d.parseCallPath(dst, req.Path)
	if err != nil {
		err = fmt.Errorf("parse rpc request(%d) failed, src:%q, dst:%q, transit:%q, path:%q, %s", req.CorrId, src, dst, transit, req.Path, err)
		go d.reply(src, transit, req.CorrId, nil, err)
		return err
	}

	switch cp.Category {
	case callpath.Service:
		go func() {
			retsRV, err := CallService(d.servCtx, cp.Plugin, cp.Method, req.Args)
			if err != nil {
				log.Errorf(d.servCtx, "rpc request(%d) service plugin:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", req.CorrId, cp.Plugin, cp.Method, src, dst, transit, req.Path, err)
			} else {
				log.Debugf(d.servCtx, "rpc request(%d) service plugin:%q, method:%q calls finished, src:%q, dst:%q, transit:%q, path:%q", req.CorrId, cp.Plugin, cp.Method, src, dst, transit, req.Path)
			}
			d.reply(src, transit, req.CorrId, retsRV, nil)
		}()

		return nil

	case callpath.Runtime:
		asyncRet, err := CallRuntime(d.servCtx, cp.EntityId, cp.Plugin, cp.Method, req.Args)
		if err != nil {
			log.Errorf(d.servCtx, "rpc request(%d) entity:%q, runtime plugin:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", req.CorrId, cp.EntityId, cp.Plugin, cp.Method, src, dst, transit, req.Path, err)
			go d.reply(src, transit, req.CorrId, nil, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(d.servCtx)
			if !ret.OK() {
				log.Errorf(d.servCtx, "rpc request(%d) entity:%q, runtime plugin:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", req.CorrId, cp.EntityId, cp.Plugin, cp.Method, src, dst, transit, req.Path, ret.Error)
				d.reply(src, transit, req.CorrId, nil, ret.Error)
			} else {
				log.Debugf(d.servCtx, "rpc request(%d) entity:%q, runtime plugin:%q, method:%q calls finished, src:%q, dst:%q, transit:%q, path:%q", req.CorrId, cp.EntityId, cp.Plugin, cp.Method, src, dst, transit, req.Path)
				d.reply(src, transit, req.CorrId, ret.Value.([]reflect.Value), nil)
			}
		}()

		return nil

	case callpath.Entity:
		asyncRet, err := CallEntity(d.servCtx, cp.EntityId, cp.Component, cp.Method, req.Args)
		if err != nil {
			log.Errorf(d.servCtx, "rpc request(%d) entity:%q, component:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", req.CorrId, cp.EntityId, cp.Component, cp.Method, src, dst, transit, req.Path, err)
			go d.reply(src, transit, req.CorrId, nil, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(d.servCtx)
			if !ret.OK() {
				log.Errorf(d.servCtx, "rpc request(%d) entity:%q, component:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", req.CorrId, cp.EntityId, cp.Component, cp.Method, src, dst, transit, req.Path, ret.Error)
				d.reply(src, transit, req.CorrId, nil, ret.Error)
			} else {
				log.Debugf(d.servCtx, "rpc request(%d) entity:%q, component:%q, method:%q calls finished, src:%q, dst:%q, transit:%q, path:%q", req.CorrId, cp.EntityId, cp.Component, cp.Method, src, dst, transit, req.Path)
				d.reply(src, transit, req.CorrId, ret.Value.([]reflect.Value), nil)
			}
		}()

		return nil
	}

	return nil
}

func (d *_ForwardInDispatcher) reply(src, transit string, corrId int64, retsRV []reflect.Value, retErr error) {
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

	bs, err := gap.Marshal(msg)
	if err != nil {
		log.Errorf(d.servCtx, "rpc reply(%d) to src:%q failed, %s", corrId, src, err)
		return
	}
	defer bs.Release()

	forwardMsg := &gap.MsgForward{
		Dst:       src,
		CorrId:    corrId,
		TransId:   msg.MsgId(),
		TransData: bs.Data(),
	}

	err = d.dist.SendMsg(transit, forwardMsg)
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

func (d *_ForwardInDispatcher) parseCallPath(dst, path string) (callpath.CallPath, error) {
	cp, err := callpath.Parse(path)
	if err != nil {
		return callpath.CallPath{}, err
	}

	if cp.EntityId.IsNil() {
		cp.EntityId = uid.From(dst)
	}

	return cp, nil
}
