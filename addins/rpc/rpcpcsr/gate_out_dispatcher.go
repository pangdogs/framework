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
	"time"

	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/addins/gate"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/variant"
	"go.uber.org/zap"
)

func (p *_GateProcessor) handleServiceMsg(topic string, mp gap.MsgPacket) {
	switch mp.Head.MsgId {
	case gap.MsgId_Forward:
		req := mp.Body.(*gap.MsgForward)

		// 只支持来源于服务域的转出消息
		if !p.dsvc.NodeDetails().DomainRoot.Contains(mp.Head.Src.Addr) || !gate.ClientDetails.DomainRoot.Contains(req.Dst) {
			return
		}

		p.acceptOutbound(mp.Head.Src, req)
	}

}

func (p *_GateProcessor) acceptOutbound(src gap.Origin, req *gap.MsgForward) {
	// 目标为单播地址，向对端发送消息
	if entityId, ok := gate.ClientDetails.DomainUnicast.Relative(req.Dst); ok {
		mapping, ok := p.router.Lookup(uid.From(entityId))
		if !ok {
			p.finishOutbound(src.Addr, req.Dst, req.CorrId, ErrSessionNotFound, req.TransId == gap.MsgId_RPC_Request)
			return
		}

		mpBuf, err := p.encoder.Encode(
			gap.Origin{Svc: p.svcCtx.Name(), Addr: p.dsvc.NodeDetails().LocalAddr, Timestamp: time.Now().UnixMilli()},
			0,
			&gap.SerializedMsg{Id: req.TransId, Data: req.TransData},
		)
		if err != nil {
			p.finishOutbound(src.Addr, req.Dst, req.CorrId, err, req.TransId == gap.MsgId_RPC_Request)
			return
		}
		defer mpBuf.Release()

		if err := mapping.Session().DataIO().Send(mpBuf.Payload()); err != nil {
			p.finishOutbound(src.Addr, req.Dst, req.CorrId, err, req.TransId == gap.MsgId_RPC_Request)
			return
		}

		p.finishOutbound(src.Addr, req.Dst, req.CorrId, nil, req.TransId == gap.MsgId_RPC_Request)
		return
	}

	// 目标为组播地址，向分组发送消息
	if gate.ClientDetails.DomainMulticast.Contains(req.Dst) {
		group, ok := p.router.GetGroupByAddr(p.svcCtx, req.Dst)
		if !ok {
			p.finishOutbound(src.Addr, req.Dst, req.CorrId, ErrGroupNotFound, req.TransId == gap.MsgId_RPC_Request)
			return
		}

		mpBuf, err := p.encoder.Encode(
			gap.Origin{Svc: p.svcCtx.Name(), Addr: p.dsvc.NodeDetails().LocalAddr, Timestamp: time.Now().UnixMilli()},
			0,
			&gap.SerializedMsg{Id: req.TransId, Data: req.TransData},
		)
		if err != nil {
			p.finishOutbound(src.Addr, req.Dst, req.CorrId, err, req.TransId == gap.MsgId_RPC_Request)
			return
		}
		defer mpBuf.Release()

		if err := group.DataIO().Send(mpBuf.Payload()); err != nil {
			p.finishOutbound(src.Addr, req.Dst, req.CorrId, err, req.TransId == gap.MsgId_RPC_Request)
			return
		}

		p.finishOutbound(src.Addr, req.Dst, req.CorrId, nil, req.TransId == gap.MsgId_RPC_Request)
		return
	}

	// 目的地址错误
	p.finishOutbound(src.Addr, req.Dst, req.CorrId, ErrIncorrectDestAddress, req.TransId == gap.MsgId_RPC_Request)
}

func (p *_GateProcessor) finishOutbound(src, dst string, corrId int64, err error, replyReject bool) {
	if err == nil {
		log.L(p.svcCtx).Debug("outbound rpc request/notify/reply forwarded",
			zap.String("src", src),
			zap.String("dst", dst),
			zap.Int64("corr_id", corrId))
	} else {
		log.L(p.svcCtx).Error("outbound rpc request/notify/reply forwarding failed",
			zap.String("src", src),
			zap.String("dst", dst),
			zap.Int64("corr_id", corrId),
			zap.Error(err))
		if replyReject {
			p.rejectOutbound(src, corrId, err)
		}
	}
}

func (p *_GateProcessor) rejectOutbound(src string, corrId int64, rejectedErr error) {
	if corrId == 0 || rejectedErr == nil {
		return
	}

	msg := &gap.MsgRPCReply{
		CorrId: corrId,
		Error:  *variant.NewError(rejectedErr),
	}

	err := p.dsvc.Send(src, msg)
	if err != nil {
		log.L(p.svcCtx).Error("send outbound rpc rejected reply failed",
			zap.String("src", src),
			zap.Int64("corr_id", corrId),
			zap.NamedError("rejected_err", rejectedErr),
			zap.Error(err))
		return
	}

	log.L(p.svcCtx).Debug("outbound rpc rejected reply sent",
		zap.String("src", src),
		zap.Int64("corr_id", corrId),
		zap.NamedError("rejected_err", rejectedErr))
}
