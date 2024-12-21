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
	"git.golaxy.org/framework/addins/dentq"
	"git.golaxy.org/framework/addins/gate"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/variant"
	"slices"
)

func (p *_GateProcessor) handleSessionChanged(session gate.ISession, curState gate.SessionState, lastState gate.SessionState) {
	switch curState {
	case gate.SessionState_Confirmed:
		err := session.GetSettings().
			RecvDataHandler(append(session.GetSettings().CurrRecvDataHandler, p.handleRecvData)).
			Change()
		if err != nil {
			log.Errorf(p.svcCtx, "change session %q settings failed, %s", session.GetId(), err)
			return
		}
	}
}

func (p *_GateProcessor) handleRecvData(session gate.ISession, data []byte) error {
	mp, err := p.decoder.Decode(data)
	if err != nil {
		return err
	}

	switch mp.Head.MsgId {
	case gap.MsgId_Forward:
		return p.acceptInbound(session, mp.Head.Seq, mp.Msg.(*gap.MsgForward))
	}

	return nil
}

func (p *_GateProcessor) acceptInbound(session gate.ISession, seq int64, req *gap.MsgForward) error {
	switch req.TransId {
	case gap.MsgId_RPC_Request, gap.MsgId_RPC_Reply, gap.MsgId_OnewayRPC:
		break
	default:
		return nil
	}

	entity, cliAddr, ok := p.router.LookupEntity(session.GetId())
	if !ok {
		go p.finishInbound(session, req.Dst, req.CorrId, ErrEntityNotFound)
		return ErrEntityNotFound
	}

	distEntity, ok := p.dentq.GetDistEntity(entity.GetId())
	if !ok {
		go p.finishInbound(session, req.Dst, req.CorrId, ErrDistEntityNotFound)
		return ErrDistEntityNotFound
	}

	nodeIdx := slices.IndexFunc(distEntity.Nodes, func(node dentq.Node) bool {
		return node.Service == req.Dst || node.RemoteAddr == req.Dst
	})
	if nodeIdx < 0 {
		go p.finishInbound(session, req.Dst, req.CorrId, ErrDistEntityNodeNotFound)
		return ErrDistEntityNodeNotFound
	}
	node := distEntity.Nodes[nodeIdx]

	msg := &gap.MsgForward{
		Transit:   p.dist.GetNodeDetails().LocalAddr, // 中转地址
		Dst:       entity.GetId().String(),           // 目标实体
		TransId:   req.TransId,
		TransData: req.TransData,
	}

	if err := p.dist.ForwardMsg(gate.CliDetails.DomainRoot.Path, cliAddr, node.RemoteAddr, seq, msg); err != nil {
		go p.finishInbound(session, node.RemoteAddr, req.CorrId, err)
		return err
	}

	go p.finishInbound(session, req.Dst, req.CorrId, nil)
	return nil
}

func (p *_GateProcessor) finishInbound(session gate.ISession, dst string, corrId int64, err error) {
	if err == nil {
		if corrId != 0 {
			log.Debugf(p.svcCtx, "forwarding session:%q rpc request(%d) to dst:%q finish", session.GetId(), corrId, dst)
		} else {
			log.Debugf(p.svcCtx, "forwarding session:%q rpc notify to dst:%q finish", session.GetId(), dst)
		}
	} else {
		if corrId != 0 {
			log.Errorf(p.svcCtx, "forwarding session:%q rpc request(%d) to dst:%q failed, %s", session.GetId(), corrId, dst, err)
			p.replyInbound(session, corrId, err)
		} else {
			log.Errorf(p.svcCtx, "forwarding session:%q rpc notify to dst:%q failed, %s", session.GetId(), dst, err)
		}
	}
}

func (p *_GateProcessor) replyInbound(session gate.ISession, corrId int64, retErr error) {
	if corrId == 0 || retErr == nil {
		return
	}

	msg := &gap.MsgRPCReply{
		CorrId: corrId,
		Error:  *variant.MakeError(retErr),
	}

	bs, err := p.encoder.Encode(p.svcCtx.GetName(), p.dist.GetNodeDetails().LocalAddr, 0, msg)
	if err != nil {
		log.Errorf(p.svcCtx, "rpc reply(%d) to session:%q failed, %s", corrId, session.GetId(), err)
		return
	}
	defer bs.Release()

	err = session.SendData(bs.Data())
	if err != nil {
		log.Errorf(p.svcCtx, "rpc reply(%d) to session:%q failed, %s", corrId, session.GetId(), err)
		return
	}

	log.Debugf(p.svcCtx, "rpc reply(%d) to session:%q ok", corrId, session.GetId())
}
