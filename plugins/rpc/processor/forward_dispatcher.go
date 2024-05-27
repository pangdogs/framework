package processor

import (
	"fmt"
	"git.golaxy.org/core/util/uid"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/variant"
	"git.golaxy.org/framework/plugins/gate"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/plugins/rpc/callpath"
	"git.golaxy.org/framework/plugins/rpcstack"
	"git.golaxy.org/framework/util/concurrent"
	"reflect"
)

func (p *_ForwardProcessor) handleMsg(topic string, mp gap.MsgPacket) error {
	// 只支持客户端域通信
	if !gate.CliDetails.InDomain(mp.Head.Src) {
		return nil
	}

	switch mp.Head.MsgId {
	case gap.MsgId_Forward:
		return p.acceptForward(mp.Head.Src, mp.Msg.(*gap.MsgForward))
	}

	return nil
}

func (p *_ForwardProcessor) acceptForward(src string, req *gap.MsgForward) error {
	switch req.TransId {
	case gap.MsgId_OneWayRPC:
		msg := &gap.MsgOneWayRPC{}
		if err := gap.Unmarshal(msg, req.TransData); err != nil {
			return err
		}
		return p.acceptNotify(src, req.Dst, req.Transit, msg)

	case gap.MsgId_RPC_Request:
		msg := &gap.MsgRPCRequest{}
		if err := gap.Unmarshal(msg, req.TransData); err != nil {
			go p.reply(src, req.Transit, req.CorrId, nil, err)
			return err
		}
		return p.acceptRequest(src, req.Dst, req.Transit, msg)

	case gap.MsgId_RPC_Reply:
		msg := &gap.MsgRPCReply{}
		if err := gap.Unmarshal(msg, req.TransData); err != nil {
			return err
		}
		return p.resolve(msg)
	}

	return nil
}

func (p *_ForwardProcessor) acceptNotify(src, dst, transit string, req *gap.MsgOneWayRPC) error {
	cp, err := p.parseCallPath(dst, req.Path)
	if err != nil {
		return fmt.Errorf("parse rpc notify failed, src:%q, dst:%q, transit:%q, path:%q, %s", src, dst, transit, req.Path, err)
	}

	callChain := rpcstack.CallChain{{Src: src, Transit: transit}}

	if len(p.permValidator) > 0 {
		passed, err := p.permValidator.Invoke(func(passed bool, err error) bool {
			return !passed || err != nil
		}, callChain, cp)
		if !passed && err == nil {
			err = ErrPermissionDenied
		}
		if err != nil {
			log.Errorf(p.servCtx, "rpc notify permission verification failed, src:%q, dst:%q, transit:%q, path:%q, %s", src, dst, transit, req.Path, err)
			return nil
		}
	}

	switch cp.Category {
	case callpath.Service:
		go func() {
			if _, err := CallService(p.servCtx, callChain, cp.Plugin, cp.Method, req.Args); err != nil {
				log.Errorf(p.servCtx, "rpc notify service plugin:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", cp.Plugin, cp.Method, src, dst, transit, req.Path, err)
			} else {
				log.Debugf(p.servCtx, "rpc notify service plugin:%q, method:%q calls finished, src:%q, dst:%q, transit:%q, path:%q", cp.Plugin, cp.Method, src, dst, transit, req.Path)
			}
		}()

		return nil

	case callpath.Runtime:
		asyncRet, err := CallRuntime(p.servCtx, callChain, cp.EntityId, cp.Plugin, cp.Method, req.Args)
		if err != nil {
			log.Errorf(p.servCtx, "rpc notify entity:%q, runtime plugin:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", cp.EntityId, cp.Plugin, cp.Method, src, dst, transit, req.Path, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(p.servCtx)
			if !ret.OK() {
				log.Errorf(p.servCtx, "rpc notify entity:%q, runtime plugin:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", cp.EntityId, cp.Plugin, cp.Method, src, dst, transit, req.Path, ret.Error)
			} else {
				log.Debugf(p.servCtx, "rpc notify entity:%q, runtime plugin:%q, method:%q calls finished, src:%q, dst:%q, transit:%q, path:%q,", cp.EntityId, cp.Plugin, cp.Method, src, dst, transit, req.Path)
			}
		}()

		return nil

	case callpath.Entity:
		asyncRet, err := CallEntity(p.servCtx, callChain, cp.EntityId, cp.Component, cp.Method, req.Args)
		if err != nil {
			log.Errorf(p.servCtx, "rpc notify entity:%q, component:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", cp.EntityId, cp.Component, cp.Method, src, dst, transit, req.Path, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(p.servCtx)
			if !ret.OK() {
				log.Errorf(p.servCtx, "rpc notify entity:%q, component:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", cp.EntityId, cp.Component, cp.Method, src, dst, transit, req.Path, ret.Error)
			} else {
				log.Debugf(p.servCtx, "rpc notify entity:%q, component:%q, method:%q calls finished, src:%q, dst:%q, transit:%q, path:%q", cp.EntityId, cp.Component, cp.Method, src, dst, transit, req.Path)
			}
		}()

		return nil
	}

	return nil
}

func (p *_ForwardProcessor) acceptRequest(src, dst, transit string, req *gap.MsgRPCRequest) error {
	cp, err := p.parseCallPath(dst, req.Path)
	if err != nil {
		err = fmt.Errorf("parse rpc request(%d) failed, src:%q, dst:%q, transit:%q, path:%q, %s", req.CorrId, src, dst, transit, req.Path, err)
		go p.reply(src, transit, req.CorrId, nil, err)
		return err
	}

	callChain := rpcstack.CallChain{{Src: src, Transit: transit}}

	if len(p.permValidator) > 0 {
		passed, err := p.permValidator.Invoke(func(passed bool, err error) bool {
			return !passed || err != nil
		}, callChain, cp)
		if !passed && err == nil {
			err = ErrPermissionDenied
		}
		if err != nil {
			log.Errorf(p.servCtx, "rpc request(%d) permission verification failed, src:%q, dst:%q, transit:%q, path:%q, %s", req.CorrId, src, dst, transit, req.Path, err)
			go p.reply(src, transit, req.CorrId, nil, err)
			return nil
		}
	}

	switch cp.Category {
	case callpath.Service:
		go func() {
			retsRV, err := CallService(p.servCtx, callChain, cp.Plugin, cp.Method, req.Args)
			if err != nil {
				log.Errorf(p.servCtx, "rpc request(%d) service plugin:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", req.CorrId, cp.Plugin, cp.Method, src, dst, transit, req.Path, err)
			} else {
				log.Debugf(p.servCtx, "rpc request(%d) service plugin:%q, method:%q calls finished, src:%q, dst:%q, transit:%q, path:%q", req.CorrId, cp.Plugin, cp.Method, src, dst, transit, req.Path)
			}
			p.reply(src, transit, req.CorrId, retsRV, err)
		}()

		return nil

	case callpath.Runtime:
		asyncRet, err := CallRuntime(p.servCtx, callChain, cp.EntityId, cp.Plugin, cp.Method, req.Args)
		if err != nil {
			log.Errorf(p.servCtx, "rpc request(%d) entity:%q, runtime plugin:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", req.CorrId, cp.EntityId, cp.Plugin, cp.Method, src, dst, transit, req.Path, err)
			go p.reply(src, transit, req.CorrId, nil, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(p.servCtx)
			if !ret.OK() {
				log.Errorf(p.servCtx, "rpc request(%d) entity:%q, runtime plugin:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", req.CorrId, cp.EntityId, cp.Plugin, cp.Method, src, dst, transit, req.Path, ret.Error)
				p.reply(src, transit, req.CorrId, nil, ret.Error)
			} else {
				log.Debugf(p.servCtx, "rpc request(%d) entity:%q, runtime plugin:%q, method:%q calls finished, src:%q, dst:%q, transit:%q, path:%q", req.CorrId, cp.EntityId, cp.Plugin, cp.Method, src, dst, transit, req.Path)
				p.reply(src, transit, req.CorrId, ret.Value.([]reflect.Value), nil)
			}
		}()

		return nil

	case callpath.Entity:
		asyncRet, err := CallEntity(p.servCtx, callChain, cp.EntityId, cp.Component, cp.Method, req.Args)
		if err != nil {
			log.Errorf(p.servCtx, "rpc request(%d) entity:%q, component:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", req.CorrId, cp.EntityId, cp.Component, cp.Method, src, dst, transit, req.Path, err)
			go p.reply(src, transit, req.CorrId, nil, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(p.servCtx)
			if !ret.OK() {
				log.Errorf(p.servCtx, "rpc request(%d) entity:%q, component:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", req.CorrId, cp.EntityId, cp.Component, cp.Method, src, dst, transit, req.Path, ret.Error)
				p.reply(src, transit, req.CorrId, nil, ret.Error)
			} else {
				log.Debugf(p.servCtx, "rpc request(%d) entity:%q, component:%q, method:%q calls finished, src:%q, dst:%q, transit:%q, path:%q", req.CorrId, cp.EntityId, cp.Component, cp.Method, src, dst, transit, req.Path)
				p.reply(src, transit, req.CorrId, ret.Value.([]reflect.Value), nil)
			}
		}()

		return nil
	}

	return nil
}

func (p *_ForwardProcessor) reply(src, transit string, corrId int64, retsRV []reflect.Value, retErr error) {
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

	bs, err := gap.Marshal(msg)
	if err != nil {
		log.Errorf(p.servCtx, "rpc reply(%d) to src:%q failed, %s", corrId, src, err)
		return
	}
	defer bs.Release()

	forwardMsg := &gap.MsgForward{
		Dst:       src,
		CorrId:    corrId,
		TransId:   msg.MsgId(),
		TransData: bs.Data(),
	}

	err = p.dist.SendMsg(transit, forwardMsg)
	if err != nil {
		log.Errorf(p.servCtx, "rpc reply(%d) to src:%q failed, %s", corrId, src, err)
		return
	}

	log.Debugf(p.servCtx, "rpc reply(%d) to src:%q ok", corrId, src)
}

func (p *_ForwardProcessor) resolve(reply *gap.MsgRPCReply) error {
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

func (p *_ForwardProcessor) parseCallPath(dst, path string) (callpath.CallPath, error) {
	cp, err := callpath.Parse(path)
	if err != nil {
		return callpath.CallPath{}, err
	}

	if cp.EntityId.IsNil() {
		cp.EntityId = uid.From(dst)
	}

	return cp, nil
}
