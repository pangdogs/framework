package processor

import (
	"context"
	"fmt"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/generic"
	"git.golaxy.org/core/util/types"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/variant"
	"git.golaxy.org/framework/plugins/dserv"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/plugins/rpc/callpath"
	"git.golaxy.org/framework/util/concurrent"
	"reflect"
)

// NewServiceDispatcher 创建分布式服务间的RPC分发器
func NewServiceDispatcher() IDispatcher {
	return &_ServiceDispatcher{}
}

// _ServiceDispatcher 分布式服务间的RPC分发器
type _ServiceDispatcher struct {
	servCtx service.Context
	dist    dserv.IDistService
	watcher dserv.IWatcher
}

// Init 初始化
func (d *_ServiceDispatcher) Init(ctx service.Context) {
	d.servCtx = ctx
	d.dist = dserv.Using(ctx)
	d.watcher = d.dist.WatchMsg(context.Background(), generic.MakeDelegateFunc2(d.handleMsg))

	log.Debugf(d.servCtx, "rpc dispatcher %q started", types.AnyFullName(*d))
}

// Shut 结束
func (d *_ServiceDispatcher) Shut(ctx service.Context) {
	<-d.watcher.Terminate()

	log.Debugf(d.servCtx, "rpc dispatcher %q stopped", types.AnyFullName(*d))
}

func (d *_ServiceDispatcher) handleMsg(topic string, mp gap.MsgPacket) error {
	// 只支持服务域通信
	if !d.dist.GetNodeDetails().InDomain(mp.Head.Src) {
		return nil
	}

	switch mp.Head.MsgId {
	case gap.MsgId_OneWayRPC:
		return d.acceptNotify(mp.Msg.(*gap.MsgOneWayRPC))

	case gap.MsgId_RPC_Request:
		return d.acceptRequest(mp.Head.Src, mp.Msg.(*gap.MsgRPCRequest))

	case gap.MsgId_RPC_Reply:
		return d.resolve(mp.Msg.(*gap.MsgRPCReply))
	}

	return nil
}

func (d *_ServiceDispatcher) acceptNotify(req *gap.MsgOneWayRPC) error {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		return fmt.Errorf("parse rpc notify path:%q failed, %s", req.Path, err)
	}

	switch cp.Category {
	case callpath.Service:
		go func() {
			if _, err := CallService(d.servCtx, cp.Plugin, cp.Method, req.Args); err != nil {
				log.Errorf(d.servCtx, "rpc notify service plugin:%q, method:%q calls failed, %s", cp.Plugin, cp.Method, err)
			} else {
				log.Debugf(d.servCtx, "rpc notify service plugin:%q, method:%q calls finished", cp.Plugin, cp.Method)
			}
			return
		}()

		return nil

	case callpath.Runtime:
		asyncRet, err := CallRuntime(d.servCtx, cp.EntityId, cp.Plugin, cp.Method, req.Args)
		if err != nil {
			log.Errorf(d.servCtx, "rpc notify entity:%q, runtime plugin:%q, method:%q calls failed, %s", cp.EntityId, cp.Plugin, cp.Method, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(d.servCtx)
			if !ret.OK() {
				log.Errorf(d.servCtx, "rpc notify entity:%q, runtime plugin:%q, method:%q calls failed, %s", cp.EntityId, cp.Plugin, cp.Method, ret.Error)
			} else {
				log.Debugf(d.servCtx, "rpc notify entity:%q, runtime plugin:%q, method:%q calls finished", cp.EntityId, cp.Plugin, cp.Method)
			}
		}()

		return nil

	case callpath.Entity:
		asyncRet, err := CallEntity(d.servCtx, cp.EntityId, cp.Component, cp.Method, req.Args)
		if err != nil {
			log.Errorf(d.servCtx, "rpc notify entity:%q, component:%q, method:%q calls failed, %s", cp.EntityId, cp.Component, cp.Method, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(d.servCtx)
			if !ret.OK() {
				log.Errorf(d.servCtx, "rpc notify entity:%q, component:%q, method:%q calls failed, %s", cp.EntityId, cp.Component, cp.Method, ret.Error)
			} else {
				log.Debugf(d.servCtx, "rpc notify entity:%q, component:%q, method:%q calls finished", cp.EntityId, cp.Component, cp.Method)
			}
		}()

		return nil
	}

	return nil
}

func (d *_ServiceDispatcher) acceptRequest(src string, req *gap.MsgRPCRequest) error {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		err = fmt.Errorf("parse rpc request(%d) path %q failed, %s", req.CorrId, req.Path, err)
		go d.reply(src, req.CorrId, nil, err)
		return err
	}

	switch cp.Category {
	case callpath.Service:
		go func() {
			retsRV, err := CallService(d.servCtx, cp.Plugin, cp.Method, req.Args)
			if err != nil {
				log.Errorf(d.servCtx, "rpc request(%d) service plugin:%q, method:%q calls failed, %s", req.CorrId, cp.Plugin, cp.Method, err)
			} else {
				log.Debugf(d.servCtx, "rpc request(%d) service plugin:%q, method:%q calls finished", req.CorrId, cp.Plugin, cp.Method)
			}
			d.reply(src, req.CorrId, retsRV, err)
		}()

		return nil

	case callpath.Runtime:
		asyncRet, err := CallRuntime(d.servCtx, cp.EntityId, cp.Plugin, cp.Method, req.Args)
		if err != nil {
			log.Errorf(d.servCtx, "rpc request(%d) entity:%q, runtime plugin:%q, method:%q calls failed, %s", req.CorrId, cp.EntityId, cp.Plugin, cp.Method, err)
			go d.reply(src, req.CorrId, nil, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(d.servCtx)
			if !ret.OK() {
				log.Errorf(d.servCtx, "rpc request(%d) entity:%q, runtime plugin:%q, method:%q calls failed, %s", req.CorrId, cp.EntityId, cp.Plugin, cp.Method, ret.Error)
				d.reply(src, req.CorrId, nil, ret.Error)
			} else {
				log.Debugf(d.servCtx, "rpc request(%d) entity:%q, runtime plugin:%q, method:%q calls finished", req.CorrId, cp.EntityId, cp.Plugin, cp.Method)
				d.reply(src, req.CorrId, ret.Value.([]reflect.Value), nil)
			}
		}()

		return nil

	case callpath.Entity:
		asyncRet, err := CallEntity(d.servCtx, cp.EntityId, cp.Component, cp.Method, req.Args)
		if err != nil {
			log.Errorf(d.servCtx, "rpc request(%d) entity:%q, component:%q, method:%q calls failed, %s", req.CorrId, cp.EntityId, cp.Component, cp.Method, err)
			go d.reply(src, req.CorrId, nil, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(d.servCtx)
			if !ret.OK() {
				log.Errorf(d.servCtx, "rpc request(%d) entity:%q, component:%q, method:%q calls failed, %s", req.CorrId, cp.EntityId, cp.Component, cp.Method, ret.Error)
				d.reply(src, req.CorrId, nil, ret.Error)
			} else {
				log.Debugf(d.servCtx, "rpc request(%d) entity:%q, component:%q, method:%q calls finished", req.CorrId, cp.EntityId, cp.Component, cp.Method)
				d.reply(src, req.CorrId, ret.Value.([]reflect.Value), nil)
			}
		}()

		return nil
	}

	return nil
}

func (d *_ServiceDispatcher) reply(src string, corrId int64, retsRV []reflect.Value, retErr error) {
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

	err = d.dist.SendMsg(src, msg)
	if err != nil {
		log.Errorf(d.servCtx, "rpc reply(%d) to src:%q failed, %s", corrId, src, err)
		return
	}

	log.Debugf(d.servCtx, "rpc reply(%d) to src:%q ok", corrId, src)
}

func (d *_ServiceDispatcher) resolve(reply *gap.MsgRPCReply) error {
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
