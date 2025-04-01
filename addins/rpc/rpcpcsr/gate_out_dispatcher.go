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
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/addins/gate"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/addins/router"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/variant"
	"time"
)

func (p *_GateProcessor) handleMsg(topic string, mp gap.MsgPacket) error {
	switch mp.Head.MsgId {
	case gap.MsgId_Forward:
		req := mp.Msg.(*gap.MsgForward)

		// 只支持来源于服务域的转出消息
		if !p.dist.GetNodeDetails().DomainRoot.Contains(mp.Head.Src.Addr) || !gate.CliDetails.DomainRoot.Contains(req.Dst) {
			return nil
		}

		p.acceptOutbound(mp.Head.Src, req)
	}

	return nil
}

func (p *_GateProcessor) acceptOutbound(src gap.Origin, req *gap.MsgForward) {
	// 目标为单播地址，为了保持消息时序，在实体线程中，向对端发送消息
	entId, ok := gate.CliDetails.DomainUnicast.Relative(req.Dst)
	if ok {
		asyncRet := p.svcCtx.CallAsync(uid.From(entId), func(entity ec.Entity, _ ...any) async.Ret {
			session, ok := p.router.LookupSession(entity.GetId())
			if !ok {
				return async.MakeRet(nil, ErrSessionNotFound)
			}

			mpBuf, err := p.encoder.Encode(
				gap.Origin{Svc: p.svcCtx.GetName(), Addr: p.dist.GetNodeDetails().LocalAddr, Timestamp: time.Now().UnixMilli()},
				0,
				&gap.SerializedMsg{Id: req.TransId, Data: req.TransData},
			)
			if err != nil {
				return async.MakeRet(nil, err)
			}
			defer mpBuf.Release()

			err = session.SendData(mpBuf.Data())
			if err != nil {
				return async.MakeRet(nil, err)
			}

			return async.MakeRet(nil, nil)
		})
		go func() { p.finishOutbound(src, req, (<-asyncRet).Error) }()
		return
	}

	// 目标为广播地址，遍历包含实体的所有分组，向每个分组发送消息
	entId, ok = gate.CliDetails.DomainBroadcast.Relative(req.Dst)
	if ok {
		p.router.EachGroups(p.svcCtx, uid.From(entId), func(group router.IGroup) {
			mpBuf, err := p.encoder.Encode(
				gap.Origin{Svc: p.svcCtx.GetName(), Addr: p.dist.GetNodeDetails().LocalAddr, Timestamp: time.Now().UnixMilli()},
				0,
				&gap.SerializedMsg{Id: req.TransId, Data: req.TransData},
			)
			if err != nil {
				go p.finishOutbound(src, req, err)
				return
			}

			// 为了保持消息时序，使用分组发送数据的channel
			select {
			case group.SendDataChan() <- mpBuf:
				go p.finishOutbound(src, req, nil)
			default:
				mpBuf.Release()
				go p.finishOutbound(src, req, ErrGroupChanIsFull)
			}
		})
		return
	}

	// 目标为组播地址，向分组发送消息
	if gate.CliDetails.DomainMulticast.Contains(req.Dst) {
		group, ok := p.router.GetGroupByAddr(p.svcCtx, req.Dst)
		if !ok {
			go p.finishOutbound(src, req, ErrGroupNotFound)
			return
		}

		mpBuf, err := p.encoder.Encode(
			gap.Origin{Svc: p.svcCtx.GetName(), Addr: p.dist.GetNodeDetails().LocalAddr, Timestamp: time.Now().UnixMilli()},
			0,
			&gap.SerializedMsg{Id: req.TransId, Data: req.TransData},
		)
		if err != nil {
			go p.finishOutbound(src, req, err)
			return
		}

		// 为了保持消息时序，使用分组发送数据的channel
		select {
		case group.SendDataChan() <- mpBuf:
			go p.finishOutbound(src, req, nil)
		default:
			mpBuf.Release()
			go p.finishOutbound(src, req, ErrGroupChanIsFull)
		}
		return
	}

	// 目的地址错误
	go p.finishOutbound(src, req, ErrIncorrectDestAddress)
}

func (p *_GateProcessor) finishOutbound(src gap.Origin, req *gap.MsgForward, err error) {
	if err == nil {
		if req.CorrId != 0 {
			log.Debugf(p.svcCtx, "outbound forwarding src:%q rpc request(%d) to remote:%q finish", src.Addr, req.CorrId, req.Dst)
		} else {
			log.Debugf(p.svcCtx, "outbound forwarding src:%q rpc notify to remote:%q finish", src.Addr, req.Dst)
		}
	} else {
		if req.CorrId != 0 {
			log.Errorf(p.svcCtx, "outbound forwarding src:%q rpc request(%d) to remote:%q failed, %s", src.Addr, req.CorrId, req.Dst, err)
			p.replyOutboundFailed(src, req.CorrId, err)
		} else {
			log.Errorf(p.svcCtx, "outbound forwarding src:%q rpc notify to remote:%q failed, %s", src.Addr, req.Dst, err)
		}
	}
}

func (p *_GateProcessor) replyOutboundFailed(src gap.Origin, corrId int64, retErr error) {
	if corrId == 0 || retErr == nil {
		return
	}

	msg := &gap.MsgRPCReply{
		CorrId: corrId,
		Error:  *variant.MakeError(retErr),
	}

	err := p.dist.SendMsg(src.Addr, msg)
	if err != nil {
		log.Errorf(p.svcCtx, "rpc reply(%d) outbound failed to src:%q failed, %s", corrId, src.Addr, err)
		return
	}

	log.Debugf(p.svcCtx, "rpc reply(%d) outbound failed to src:%q ok", corrId, src.Addr)
}
