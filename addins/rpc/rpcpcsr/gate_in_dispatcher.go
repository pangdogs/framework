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
	"slices"
	"time"

	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/addins/dent"
	"git.golaxy.org/framework/addins/gate"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/variant"
	"go.uber.org/zap"
)

func (p *_GateProcessor) handleSessionEstablished(session gate.ISession) {
	err := session.DataIO().Listen(p.stoppingCtx, generic.CastDelegateVoid2(p.handleSessionData))
	if err != nil {
		log.L(p.svcCtx).Error("listen session data failed",
			zap.String("session_id", session.Id().String()),
			zap.Error(err))
		return
	}
	log.L(p.svcCtx).Debug("listen session data started", zap.String("session_id", session.Id().String()))
}

func (p *_GateProcessor) handleSessionData(session gate.ISession, data []byte) {
	mp, err := p.decoder.Decode(data)
	if err != nil {
		log.L(p.svcCtx).Error("decode session data failed",
			zap.String("session_id", session.Id().String()),
			zap.Error(err))
		return
	}

	switch mp.Head.MsgId {
	case gap.MsgId_Forward:
		p.acceptInbound(session, mp.Head.Src.Timestamp, mp.Msg.(*gap.MsgForward))
	}
}

func (p *_GateProcessor) acceptInbound(session gate.ISession, timestamp int64, req *gap.MsgForward) {
	switch req.TransId {
	case gap.MsgId_RPC_Request, gap.MsgId_OnewayRPC, gap.MsgId_RPC_Reply:
		break
	default:
		return
	}

	mapping, ok := p.router.Lookup(session.Id())
	if !ok {
		p.finishInbound(session, "", req.Dst, req.CorrId, ErrEntityNotFound, req.TransId == gap.MsgId_RPC_Request)
		return
	}

	distEntity, ok := p.dentq.GetDistEntity(mapping.Entity().Id())
	if !ok {
		p.finishInbound(session, mapping.ClientAddr(), req.Dst, req.CorrId, ErrDistEntityNotFound, req.TransId == gap.MsgId_RPC_Request)
		return
	}

	nodeIdx := slices.IndexFunc(distEntity.Nodes, func(node dent.Node) bool {
		return node.Service == req.Dst || node.RemoteAddr == req.Dst
	})
	if nodeIdx < 0 {
		p.finishInbound(session, mapping.ClientAddr(), req.Dst, req.CorrId, ErrDistEntityNodeNotFound, req.TransId == gap.MsgId_RPC_Request)
		return
	}
	node := distEntity.Nodes[nodeIdx]

	msg := &gap.MsgForward{
		Src: gap.Origin{
			Svc:       gate.ClientDetails.DomainRoot.Path,
			Addr:      mapping.ClientAddr(),
			Timestamp: timestamp,
		},
		Dst:       mapping.Entity().Id().String(), // 目标实体
		CorrId:    req.CorrId,
		TransId:   req.TransId,
		TransData: req.TransData,
	}

	if err := p.dsvc.Send(node.RemoteAddr, msg); err != nil {
		p.finishInbound(session, mapping.ClientAddr(), node.RemoteAddr, req.CorrId, err, req.TransId == gap.MsgId_RPC_Request)
		return
	}

	p.finishInbound(session, mapping.ClientAddr(), req.Dst, req.CorrId, nil, req.TransId == gap.MsgId_RPC_Request)
}

func (p *_GateProcessor) finishInbound(session gate.ISession, src, dst string, corrId int64, err error, replyReject bool) {
	if err == nil {
		log.L(p.svcCtx).Debug("inbound rpc request/notify/reply forwarded",
			zap.String("session_id", session.Id().String()),
			zap.String("local", session.NetAddr().Local.String()),
			zap.String("remote", session.NetAddr().Remote.String()),
			zap.String("src", src),
			zap.String("dst", dst),
			zap.Int64("corr_id", corrId))
	} else {
		log.L(p.svcCtx).Error("inbound rpc request/notify/reply forwarding failed",
			zap.String("session_id", session.Id().String()),
			zap.String("local", session.NetAddr().Local.String()),
			zap.String("remote", session.NetAddr().Remote.String()),
			zap.String("src", src),
			zap.String("dst", dst),
			zap.Int64("corr_id", corrId),
			zap.Error(err))
		if replyReject {
			p.rejectInbound(session, corrId, err)
		}
	}
}

func (p *_GateProcessor) rejectInbound(session gate.ISession, corrId int64, rejectedErr error) {
	if corrId == 0 || rejectedErr == nil {
		return
	}

	mpBuf, err := p.encoder.Encode(
		gap.Origin{Svc: p.svcCtx.Name(), Addr: p.dsvc.NodeDetails().LocalAddr, Timestamp: time.Now().UnixMilli()},
		0,
		&gap.MsgRPCReply{CorrId: corrId, Error: *variant.NewError(rejectedErr)},
	)
	if err != nil {
		log.L(p.svcCtx).Error("encode inbound rpc rejected reply failed",
			zap.String("session_id", session.Id().String()),
			zap.Int64("corr_id", corrId),
			zap.NamedError("rejected_err", rejectedErr),
			zap.Error(err))
		return
	}
	defer mpBuf.Release()

	if err := session.DataIO().Send(mpBuf.Payload()); err != nil {
		log.L(p.svcCtx).Error("send inbound rpc rejected reply failed",
			zap.String("session_id", session.Id().String()),
			zap.Int64("corr_id", corrId),
			zap.NamedError("rejected_err", rejectedErr),
			zap.Error(err))
		return
	}

	log.L(p.svcCtx).Debug("inbound rpc rejected reply sent",
		zap.String("session_id", session.Id().String()),
		zap.Int64("corr_id", corrId),
		zap.NamedError("rejected_err", rejectedErr))
}
