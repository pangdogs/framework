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
	"git.golaxy.org/framework/utils/correlation"
	"go.uber.org/zap"
)

// handleSessionEstablished 为新会话注册原始数据监听器。
func (p *_GateProcessor) handleSessionEstablished(session gate.ISession) {
	err := session.DataIO().Listen(p.scope.Context(), generic.CastDelegateVoid2(p.handleSessionData))
	if err != nil {
		log.L(p.svcCtx).Error("listen session data failed",
			zap.String("session_id", session.ID().String()),
			zap.Error(err))
		return
	}
	log.L(p.svcCtx).Debug("listen session data started", zap.String("session_id", session.ID().String()))
}

// handleSessionData 解码客户端 GAP 数据并接收入站转发消息。
func (p *_GateProcessor) handleSessionData(session gate.ISession, data []byte) {
	mp, err := p.decoder.Decode(data)
	if err != nil {
		log.L(p.svcCtx).Error("decode session data failed",
			zap.String("session_id", session.ID().String()),
			zap.Error(err))
		return
	}

	switch mp.Head.MsgID {
	case gap.MsgID_Forward:
		p.acceptInbound(session, mp.Head.Src.Timestamp, mp.Body.(*gap.MsgForward))
	}
}

// acceptInbound 根据会话映射和分布式实体位置，将客户端 RPC 转发到目标服务节点。
func (p *_GateProcessor) acceptInbound(session gate.ISession, timestamp int64, req *gap.MsgForward) {
	switch req.TransID {
	case gap.MsgID_RPC_Request, gap.MsgID_OnewayRPC, gap.MsgID_RPC_Reply:
		break
	default:
		return
	}

	mapping, ok := p.router.Lookup(session.ID())
	if !ok {
		p.finishInbound(session, "", req.Dst, req.CorrID, ErrEntityNotFound, req.TransID == gap.MsgID_RPC_Request)
		return
	}

	distEntity, ok := p.dentq.GetDistEntity(mapping.Entity().ID())
	if !ok {
		p.finishInbound(session, mapping.ClientAddr(), req.Dst, req.CorrID, ErrDistEntityNotFound, req.TransID == gap.MsgID_RPC_Request)
		return
	}

	nodeIdx := slices.IndexFunc(distEntity.Nodes, func(node dent.Node) bool {
		return node.Service == req.Dst || node.RemoteAddr == req.Dst
	})
	if nodeIdx < 0 {
		p.finishInbound(session, mapping.ClientAddr(), req.Dst, req.CorrID, ErrDistEntityNodeNotFound, req.TransID == gap.MsgID_RPC_Request)
		return
	}
	node := distEntity.Nodes[nodeIdx]

	msg := &gap.MsgForward{
		Src: gap.Origin{
			Svc:       gate.ClientDetails.DomainRoot.Path,
			Addr:      mapping.ClientAddr(),
			Timestamp: timestamp,
		},
		Dst:       mapping.Entity().ID().String(), // 目标实体
		CorrID:    req.CorrID,
		TransID:   req.TransID,
		TransData: req.TransData,
	}

	if err := p.dsvc.Send(node.RemoteAddr, msg); err != nil {
		p.finishInbound(session, mapping.ClientAddr(), node.RemoteAddr, req.CorrID, err, req.TransID == gap.MsgID_RPC_Request)
		return
	}

	p.finishInbound(session, mapping.ClientAddr(), req.Dst, req.CorrID, nil, req.TransID == gap.MsgID_RPC_Request)
}

// finishInbound 记录转发结果，并在请求转发失败时向客户端回复拒绝错误。
func (p *_GateProcessor) finishInbound(session gate.ISession, src, dst string, corrID correlation.ID, err error, replyReject bool) {
	if err == nil {
		log.L(p.svcCtx).Debug("inbound rpc request/notify/reply forwarded",
			zap.String("session_id", session.ID().String()),
			zap.String("local", session.NetAddr().Local.String()),
			zap.String("remote", session.NetAddr().Remote.String()),
			zap.String("src", src),
			zap.String("dst", dst),
			zap.Uint64("corr_id", uint64(corrID)))
	} else {
		log.L(p.svcCtx).Error("inbound rpc request/notify/reply forwarding failed",
			zap.String("session_id", session.ID().String()),
			zap.String("local", session.NetAddr().Local.String()),
			zap.String("remote", session.NetAddr().Remote.String()),
			zap.String("src", src),
			zap.String("dst", dst),
			zap.Uint64("corr_id", uint64(corrID)),
			zap.Error(err))
		if replyReject {
			p.rejectInbound(session, corrID, err)
		}
	}
}

// rejectInbound 将入站请求失败转换为 RPC 响应并直接发送给客户端。
func (p *_GateProcessor) rejectInbound(session gate.ISession, corrID correlation.ID, rejectedErr error) {
	if corrID == 0 || rejectedErr == nil {
		return
	}

	mpBuf, err := p.encoder.Encode(
		gap.Origin{Svc: p.svcCtx.Name(), Addr: p.dsvc.NodeDetails().LocalAddr, Timestamp: time.Now().UnixMilli()},
		0,
		&gap.MsgRPCReply{CorrID: corrID, Error: *variant.NewError(rejectedErr)},
	)
	if err != nil {
		log.L(p.svcCtx).Error("encode inbound rpc rejected reply failed",
			zap.String("session_id", session.ID().String()),
			zap.Uint64("corr_id", uint64(corrID)),
			zap.NamedError("rejected_err", rejectedErr),
			zap.Error(err))
		return
	}
	defer mpBuf.Release()

	if err := session.DataIO().Send(mpBuf.Payload()); err != nil {
		log.L(p.svcCtx).Error("send inbound rpc rejected reply failed",
			zap.String("session_id", session.ID().String()),
			zap.Uint64("corr_id", uint64(corrID)),
			zap.NamedError("rejected_err", rejectedErr),
			zap.Error(err))
		return
	}

	log.L(p.svcCtx).Debug("inbound rpc rejected reply sent",
		zap.String("session_id", session.ID().String()),
		zap.Uint64("corr_id", uint64(corrID)),
		zap.NamedError("rejected_err", rejectedErr))
}
