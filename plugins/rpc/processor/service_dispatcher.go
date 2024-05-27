package processor

import (
	"fmt"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/variant"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/plugins/rpc/callpath"
	"git.golaxy.org/framework/plugins/rpcstack"
	"git.golaxy.org/framework/util/concurrent"
	"reflect"
)

func (p *_ServiceProcessor) handleMsg(topic string, mp gap.MsgPacket) error {
	// 只支持服务域通信
	if !p.dist.GetNodeDetails().InDomain(mp.Head.Src) {
		return nil
	}

	switch mp.Head.MsgId {
	case gap.MsgId_OneWayRPC:
		return p.acceptNotify(mp.Head.Src, mp.Msg.(*gap.MsgOneWayRPC))

	case gap.MsgId_RPC_Request:
		return p.acceptRequest(mp.Head.Src, mp.Msg.(*gap.MsgRPCRequest))

	case gap.MsgId_RPC_Reply:
		return p.resolve(mp.Msg.(*gap.MsgRPCReply))
	}

	return nil
}

func (p *_ServiceProcessor) acceptNotify(src string, req *gap.MsgOneWayRPC) error {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		return fmt.Errorf("parse rpc notify path:%q failed, %s", req.Path, err)
	}

	if cp.ExcludeSrc && src == p.dist.GetNodeDetails().LocalAddr {
		return nil
	}

	callChain := append(req.CallChain, rpcstack.Call{Src: src})

	if len(p.permValidator) > 0 {
		passed, err := p.permValidator.Invoke(func(passed bool, err error) bool {
			return !passed || err != nil
		}, callChain, cp)
		if !passed && err == nil {
			err = ErrPermissionDenied
		}
		if err != nil {
			log.Errorf(p.servCtx, "rpc notify permission verification failed, src:%q, path:%q, %s", src, req.Path, err)
			return nil
		}
	}

	switch cp.Category {
	case callpath.Service:
		go func() {
			if _, err := CallService(p.servCtx, callChain, cp.Plugin, cp.Method, req.Args); err != nil {
				log.Errorf(p.servCtx, "rpc notify service plugin:%q, method:%q calls failed, %s", cp.Plugin, cp.Method, err)
			} else {
				log.Debugf(p.servCtx, "rpc notify service plugin:%q, method:%q calls finished", cp.Plugin, cp.Method)
			}
			return
		}()

		return nil

	case callpath.Runtime:
		asyncRet, err := CallRuntime(p.servCtx, callChain, cp.EntityId, cp.Plugin, cp.Method, req.Args)
		if err != nil {
			log.Errorf(p.servCtx, "rpc notify entity:%q, runtime plugin:%q, method:%q calls failed, %s", cp.EntityId, cp.Plugin, cp.Method, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(p.servCtx)
			if !ret.OK() {
				log.Errorf(p.servCtx, "rpc notify entity:%q, runtime plugin:%q, method:%q calls failed, %s", cp.EntityId, cp.Plugin, cp.Method, ret.Error)
			} else {
				log.Debugf(p.servCtx, "rpc notify entity:%q, runtime plugin:%q, method:%q calls finished", cp.EntityId, cp.Plugin, cp.Method)
			}
		}()

		return nil

	case callpath.Entity:
		asyncRet, err := CallEntity(p.servCtx, callChain, cp.EntityId, cp.Component, cp.Method, req.Args)
		if err != nil {
			log.Errorf(p.servCtx, "rpc notify entity:%q, component:%q, method:%q calls failed, %s", cp.EntityId, cp.Component, cp.Method, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(p.servCtx)
			if !ret.OK() {
				log.Errorf(p.servCtx, "rpc notify entity:%q, component:%q, method:%q calls failed, %s", cp.EntityId, cp.Component, cp.Method, ret.Error)
			} else {
				log.Debugf(p.servCtx, "rpc notify entity:%q, component:%q, method:%q calls finished", cp.EntityId, cp.Component, cp.Method)
			}
		}()

		return nil
	}

	return nil
}

func (p *_ServiceProcessor) acceptRequest(src string, req *gap.MsgRPCRequest) error {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		err = fmt.Errorf("parse rpc request(%d) path %q failed, %s", req.CorrId, req.Path, err)
		go p.reply(src, req.CorrId, nil, err)
		return err
	}

	callChain := append(req.CallChain, rpcstack.Call{Src: src})

	if len(p.permValidator) > 0 {
		passed, err := p.permValidator.Invoke(func(passed bool, err error) bool {
			return !passed || err != nil
		}, callChain, cp)
		if !passed && err == nil {
			err = ErrPermissionDenied
		}
		if err != nil {
			log.Errorf(p.servCtx, "rpc request(%d) permission verification failed, src:%q, path:%q, %s", req.CorrId, src, req.Path, err)
			go p.reply(src, req.CorrId, nil, err)
			return nil
		}
	}

	switch cp.Category {
	case callpath.Service:
		go func() {
			retsRV, err := CallService(p.servCtx, callChain, cp.Plugin, cp.Method, req.Args)
			if err != nil {
				log.Errorf(p.servCtx, "rpc request(%d) service plugin:%q, method:%q calls failed, %s", req.CorrId, cp.Plugin, cp.Method, err)
			} else {
				log.Debugf(p.servCtx, "rpc request(%d) service plugin:%q, method:%q calls finished", req.CorrId, cp.Plugin, cp.Method)
			}
			p.reply(src, req.CorrId, retsRV, err)
		}()

		return nil

	case callpath.Runtime:
		asyncRet, err := CallRuntime(p.servCtx, callChain, cp.EntityId, cp.Plugin, cp.Method, req.Args)
		if err != nil {
			log.Errorf(p.servCtx, "rpc request(%d) entity:%q, runtime plugin:%q, method:%q calls failed, %s", req.CorrId, cp.EntityId, cp.Plugin, cp.Method, err)
			go p.reply(src, req.CorrId, nil, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(p.servCtx)
			if !ret.OK() {
				log.Errorf(p.servCtx, "rpc request(%d) entity:%q, runtime plugin:%q, method:%q calls failed, %s", req.CorrId, cp.EntityId, cp.Plugin, cp.Method, ret.Error)
				p.reply(src, req.CorrId, nil, ret.Error)
			} else {
				log.Debugf(p.servCtx, "rpc request(%d) entity:%q, runtime plugin:%q, method:%q calls finished", req.CorrId, cp.EntityId, cp.Plugin, cp.Method)
				p.reply(src, req.CorrId, ret.Value.([]reflect.Value), nil)
			}
		}()

		return nil

	case callpath.Entity:
		asyncRet, err := CallEntity(p.servCtx, callChain, cp.EntityId, cp.Component, cp.Method, req.Args)
		if err != nil {
			log.Errorf(p.servCtx, "rpc request(%d) entity:%q, component:%q, method:%q calls failed, %s", req.CorrId, cp.EntityId, cp.Component, cp.Method, err)
			go p.reply(src, req.CorrId, nil, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(p.servCtx)
			if !ret.OK() {
				log.Errorf(p.servCtx, "rpc request(%d) entity:%q, component:%q, method:%q calls failed, %s", req.CorrId, cp.EntityId, cp.Component, cp.Method, ret.Error)
				p.reply(src, req.CorrId, nil, ret.Error)
			} else {
				log.Debugf(p.servCtx, "rpc request(%d) entity:%q, component:%q, method:%q calls finished", req.CorrId, cp.EntityId, cp.Component, cp.Method)
				p.reply(src, req.CorrId, ret.Value.([]reflect.Value), nil)
			}
		}()

		return nil
	}

	return nil
}

func (p *_ServiceProcessor) reply(src string, corrId int64, retsRV []reflect.Value, retErr error) {
	if corrId == 0 {
		return
	}

	msg := &gap.MsgRPCReply{
		CorrId: corrId,
	}

	var err error
	msg.Rets, err = variant.MakeArray(retsRV)
	if err != nil {
		log.Errorf(p.servCtx, "rpc reply(%d) to src:%q failed, %s", corrId, src, err)
		return
	}

	if retErr != nil {
		msg.Error = *variant.MakeError(retErr)
	}

	err = p.dist.SendMsg(src, msg)
	if err != nil {
		log.Errorf(p.servCtx, "rpc reply(%d) to src:%q failed, %s", corrId, src, err)
		return
	}

	log.Debugf(p.servCtx, "rpc reply(%d) to src:%q ok", corrId, src)
}

func (p *_ServiceProcessor) resolve(reply *gap.MsgRPCReply) error {
	ret := concurrent.Ret[any]{}

	if reply.Error.OK() {
		if len(reply.Rets) > 0 {
			ret.Value = reply.Rets
		}
	} else {
		ret.Error = &reply.Error
	}

	return p.dist.GetFutures().Resolve(reply.CorrId, ret)
}
