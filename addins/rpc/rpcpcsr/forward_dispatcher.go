/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package rpcpcsr

import (
	"fmt"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/addins/gate"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/addins/rpc/callpath"
	"git.golaxy.org/framework/addins/rpcstack"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/variant"
	"time"
)

func (p *_ForwardProcessor) handleMsg(topic string, mp gap.MsgPacket) error {
	// 只支持客户端域通信
	if !gate.CliDetails.DomainRoot.Contains(mp.Head.Src) {
		return nil
	}

	switch mp.Head.MsgId {
	case gap.MsgId_Forward:
		return p.acceptForward(mp.Head.Svc, mp.Head.Src, mp.Msg.(*gap.MsgForward))
	}

	return nil
}

func (p *_ForwardProcessor) acceptForward(svc, src string, req *gap.MsgForward) error {
	switch req.TransId {
	case gap.MsgId_OnewayRPC:
		msg := &gap.MsgOnewayRPC{}
		if err := gap.Unmarshal(msg, req.TransData); err != nil {
			return err
		}
		return p.acceptNotify(svc, src, req.Dst, req.Transit, msg)

	case gap.MsgId_RPC_Request:
		msg := &gap.MsgRPCRequest{}
		if err := gap.Unmarshal(msg, req.TransData); err != nil {
			go p.reply(src, req.Transit, req.CorrId, nil, err)
			return err
		}
		return p.acceptRequest(svc, src, req.Dst, req.Transit, msg)

	case gap.MsgId_RPC_Reply:
		msg := &gap.MsgRPCReply{}
		if err := gap.Unmarshal(msg, req.TransData); err != nil {
			return err
		}
		return p.resolve(msg)
	}

	return nil
}

func (p *_ForwardProcessor) acceptNotify(svc, src, dst, transit string, req *gap.MsgOnewayRPC) error {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		return fmt.Errorf("parse rpc notify failed, src:%q, dst:%q, transit:%q, path:%q, %s", src, dst, transit, req.Path, err)
	}
	cp.Id = uid.From(dst)

	cc := rpcstack.CallChain{{Svc: svc, Addr: src, Transit: transit, Time: time.Now().UnixMilli()}}

	if len(p.permValidator) > 0 {
		passed, err := p.permValidator.SafeCall(func(passed bool, err error) bool {
			return !passed || err != nil
		}, cc, cp)
		if !passed && err == nil {
			err = ErrPermissionDenied
		}
		if err != nil {
			log.Errorf(p.svcCtx, "rpc notify permission verification failed, src:%q, dst:%q, transit:%q, path:%q, %s", src, dst, transit, req.Path, err)
			return nil
		}
	}

	switch cp.Category {
	case callpath.Service:
		go func() {
			rets, err := CallService(p.svcCtx, cc, cp.Script, cp.Method, req.Args)
			if err != nil {
				log.Errorf(p.svcCtx, "rpc notify service addIn:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", cp.Script, cp.Method, src, dst, transit, req.Path, err)
			} else {
				log.Debugf(p.svcCtx, "rpc notify service addIn:%q, method:%q calls finished, src:%q, dst:%q, transit:%q, path:%q", cp.Script, cp.Method, src, dst, transit, req.Path)
				rets.Release()
			}
		}()

		return nil

	case callpath.Runtime:
		asyncRet, err := CallRuntime(p.svcCtx, cc, cp.Id, cp.Script, cp.Method, req.Args)
		if err != nil {
			log.Errorf(p.svcCtx, "rpc notify entity:%q, runtime addIn:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", cp.Id, cp.Script, cp.Method, src, dst, transit, req.Path, err)
			return nil
		}

		go func() {
			rets, err := waitAsyncRet(p.svcCtx, asyncRet)
			if err != nil {
				log.Errorf(p.svcCtx, "rpc notify entity:%q, runtime addIn:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", cp.Id, cp.Script, cp.Method, src, dst, transit, req.Path, err)
			} else {
				log.Debugf(p.svcCtx, "rpc notify entity:%q, runtime addIn:%q, method:%q calls finished, src:%q, dst:%q, transit:%q, path:%q,", cp.Id, cp.Script, cp.Method, src, dst, transit, req.Path)
				rets.Release()
			}
		}()

		return nil

	case callpath.Entity:
		asyncRet, err := CallEntity(p.svcCtx, cc, cp.Id, cp.Script, cp.Method, req.Args)
		if err != nil {
			log.Errorf(p.svcCtx, "rpc notify entity:%q, component:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", cp.Id, cp.Script, cp.Method, src, dst, transit, req.Path, err)
			return nil
		}

		go func() {
			rets, err := waitAsyncRet(p.svcCtx, asyncRet)
			if err != nil {
				log.Errorf(p.svcCtx, "rpc notify entity:%q, component:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", cp.Id, cp.Script, cp.Method, src, dst, transit, req.Path, err)
			} else {
				log.Debugf(p.svcCtx, "rpc notify entity:%q, component:%q, method:%q calls finished, src:%q, dst:%q, transit:%q, path:%q", cp.Id, cp.Script, cp.Method, src, dst, transit, req.Path)
				rets.Release()
			}
		}()

		return nil
	}

	return nil
}

func (p *_ForwardProcessor) acceptRequest(svc, src, dst, transit string, req *gap.MsgRPCRequest) error {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		err = fmt.Errorf("parse rpc request(%d) failed, src:%q, dst:%q, transit:%q, path:%q, %s", req.CorrId, src, dst, transit, req.Path, err)
		go p.reply(src, transit, req.CorrId, nil, err)
		return err
	}
	cp.Id = uid.From(dst)

	cc := rpcstack.CallChain{{Svc: svc, Addr: src, Transit: transit, Time: time.Now().UnixMilli()}}

	if len(p.permValidator) > 0 {
		passed, err := p.permValidator.SafeCall(func(passed bool, err error) bool {
			return !passed || err != nil
		}, cc, cp)
		if !passed && err == nil {
			err = ErrPermissionDenied
		}
		if err != nil {
			log.Errorf(p.svcCtx, "rpc request(%d) permission verification failed, src:%q, dst:%q, transit:%q, path:%q, %s", req.CorrId, src, dst, transit, req.Path, err)
			go p.reply(src, transit, req.CorrId, nil, err)
			return nil
		}
	}

	switch cp.Category {
	case callpath.Service:
		go func() {
			rets, err := CallService(p.svcCtx, cc, cp.Script, cp.Method, req.Args)
			if err != nil {
				log.Errorf(p.svcCtx, "rpc request(%d) service addIn:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", req.CorrId, cp.Script, cp.Method, src, dst, transit, req.Path, err)
			} else {
				log.Debugf(p.svcCtx, "rpc request(%d) service addIn:%q, method:%q calls finished, src:%q, dst:%q, transit:%q, path:%q", req.CorrId, cp.Script, cp.Method, src, dst, transit, req.Path)
			}
			p.reply(src, transit, req.CorrId, rets, err)
		}()

		return nil

	case callpath.Runtime:
		asyncRet, err := CallRuntime(p.svcCtx, cc, cp.Id, cp.Script, cp.Method, req.Args)
		if err != nil {
			log.Errorf(p.svcCtx, "rpc request(%d) entity:%q, runtime addIn:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", req.CorrId, cp.Id, cp.Script, cp.Method, src, dst, transit, req.Path, err)
			go p.reply(src, transit, req.CorrId, nil, err)
			return nil
		}

		go func() {
			rets, err := waitAsyncRet(p.svcCtx, asyncRet)
			if err != nil {
				log.Errorf(p.svcCtx, "rpc request(%d) entity:%q, runtime addIn:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", req.CorrId, cp.Id, cp.Script, cp.Method, src, dst, transit, req.Path, err)
				p.reply(src, transit, req.CorrId, nil, err)
			} else {
				log.Debugf(p.svcCtx, "rpc request(%d) entity:%q, runtime addIn:%q, method:%q calls finished, src:%q, dst:%q, transit:%q, path:%q", req.CorrId, cp.Id, cp.Script, cp.Method, src, dst, transit, req.Path)
				p.reply(src, transit, req.CorrId, rets, nil)
			}
		}()

		return nil

	case callpath.Entity:
		asyncRet, err := CallEntity(p.svcCtx, cc, cp.Id, cp.Script, cp.Method, req.Args)
		if err != nil {
			log.Errorf(p.svcCtx, "rpc request(%d) entity:%q, component:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", req.CorrId, cp.Id, cp.Script, cp.Method, src, dst, transit, req.Path, err)
			go p.reply(src, transit, req.CorrId, nil, err)
			return nil
		}

		go func() {
			rets, err := waitAsyncRet(p.svcCtx, asyncRet)
			if err != nil {
				log.Errorf(p.svcCtx, "rpc request(%d) entity:%q, component:%q, method:%q calls failed, src:%q, dst:%q, transit:%q, path:%q, %s", req.CorrId, cp.Id, cp.Script, cp.Method, src, dst, transit, req.Path, err)
				p.reply(src, transit, req.CorrId, nil, err)
			} else {
				log.Debugf(p.svcCtx, "rpc request(%d) entity:%q, component:%q, method:%q calls finished, src:%q, dst:%q, transit:%q, path:%q", req.CorrId, cp.Id, cp.Script, cp.Method, src, dst, transit, req.Path)
				p.reply(src, transit, req.CorrId, rets, nil)
			}
		}()

		return nil
	}

	return nil
}

func (p *_ForwardProcessor) reply(src, transit string, corrId int64, rets variant.Array, retErr error) {
	defer rets.Release()

	if corrId == 0 {
		return
	}

	msg := &gap.MsgRPCReply{
		CorrId: corrId,
		Rets:   rets,
	}

	if retErr != nil {
		msg.Error = *variant.MakeError(retErr)
	}

	bs, err := gap.Marshal(msg)
	if err != nil {
		log.Errorf(p.svcCtx, "rpc reply(%d) to src:%q failed, %s", corrId, src, err)
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
		log.Errorf(p.svcCtx, "rpc reply(%d) to src:%q failed, %s", corrId, src, err)
		return
	}

	log.Debugf(p.svcCtx, "rpc reply(%d) to src:%q ok", corrId, src)
}

func (p *_ForwardProcessor) resolve(reply *gap.MsgRPCReply) error {
	ret := async.Ret{}

	if reply.Error.OK() {
		if len(reply.Rets) > 0 {
			ret.Value = reply.Rets
		}
	} else {
		ret.Error = &reply.Error
	}

	return p.dist.GetFutures().Resolve(reply.CorrId, ret)
}
