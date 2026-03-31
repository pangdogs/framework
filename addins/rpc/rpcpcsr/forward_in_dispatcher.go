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
	"time"

	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/addins/gate"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/addins/rpc/callpath"
	"git.golaxy.org/framework/addins/rpcstack"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/variant"
	"go.uber.org/zap"
)

func (p *_ForwardProcessor) handleServiceMsg(topic string, mp gap.MsgPacket) {
	switch mp.Head.MsgId {
	case gap.MsgId_Forward:
		req := mp.Msg.(*gap.MsgForward)

		// 只支持来源于客户端域的转入消息
		if !p.dsvc.NodeDetails().DomainRoot.Contains(mp.Head.Src.Addr) || !gate.ClientDetails.DomainRoot.Contains(req.Src.Addr) {
			return
		}

		p.acceptForward(mp.Head.Src, req)
	}
}

func (p *_ForwardProcessor) acceptForward(transit gap.Origin, req *gap.MsgForward) {
	switch req.TransId {
	case gap.MsgId_OnewayRPC:
		msg := &gap.MsgOnewayRPC{}
		if err := gap.Unmarshal(msg, req.TransData); err != nil {
			log.L(p.svcCtx).Error("unmarshal forwarded rpc message failed",
				zap.String("transit", transit.Addr),
				zap.String("src", req.Src.Addr),
				zap.String("dst", req.Dst),
				zap.Error(err))
			return
		}
		p.acceptNotify(transit, req.Src, req.Dst, msg)

	case gap.MsgId_RPC_Request:
		msg := &gap.MsgRPCRequest{}
		if err := gap.Unmarshal(msg, req.TransData); err != nil {
			log.L(p.svcCtx).Error("unmarshal forwarded rpc message failed",
				zap.String("transit", transit.Addr),
				zap.String("src", req.Src.Addr),
				zap.String("dst", req.Dst),
				zap.Error(err))
			p.reply(transit, req.Src, req.CorrId, nil, err)
			return
		}
		p.acceptRequest(transit, req.Src, req.Dst, msg)

	case gap.MsgId_RPC_Reply:
		msg := &gap.MsgRPCReply{}
		if err := gap.Unmarshal(msg, req.TransData); err != nil {
			log.L(p.svcCtx).Error("unmarshal forwarded rpc message failed",
				zap.String("transit", transit.Addr),
				zap.String("src", req.Src.Addr),
				zap.String("dst", req.Dst),
				zap.Error(err))
			return
		}
		p.resolveReply(transit, req.Src, msg)
	}
}

func (p *_ForwardProcessor) acceptNotify(transit, src gap.Origin, dst string, req *gap.MsgOnewayRPC) {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		log.L(p.svcCtx).Error("accept forwarded rpc notify failed",
			zap.String("transit", transit.Addr),
			zap.String("src", src.Addr),
			zap.String("dst", dst),
			zap.Error(fmt.Errorf("parse call path failed: %w", err)))
		return
	}
	cp.Id = uid.From(dst)

	cc := rpcstack.CallChain{
		{
			Svc:       src.Svc,
			Addr:      src.Addr,
			Timestamp: time.UnixMilli(src.Timestamp).Local(),
			Transit:   false,
		},
		{
			Svc:       transit.Svc,
			Addr:      transit.Addr,
			Timestamp: time.UnixMilli(transit.Timestamp).Local(),
			Transit:   true,
		},
	}

	if len(p.permValidator) > 0 {
		passed, err := p.permValidator.SafeCall(func(passed bool, err error) bool {
			return !passed || err != nil
		}, cc, cp)
		if err != nil {
			err = fmt.Errorf("%w: %w", ErrPermissionDenied, err)
		} else if !passed {
			err = ErrPermissionDenied
		}
		if err != nil {
			log.L(p.svcCtx).Error("accept forwarded rpc notify failed",
				zap.String("transit", transit.Addr),
				zap.String("src", src.Addr),
				zap.String("dst", dst),
				zap.String("call_path", cp.String()),
				zap.Error(fmt.Errorf("permission verification failed: %w", err)))
			return
		}
	}

	switch cp.TargetKind {
	case callpath.Service:
		go func() {
			rets, err := CallService(p.svcCtx, cc, cp.Script, cp.Method, req.Args)
			if err != nil {
				log.L(p.svcCtx).Error("accept forwarded rpc notify to service failed",
					zap.String("transit", transit.Addr),
					zap.String("src", src.Addr),
					zap.String("dst", dst),
					zap.String("call_path", cp.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method),
					zap.Error(err))
			} else {
				log.L(p.svcCtx).Debug("accept forwarded rpc notify to service finished",
					zap.String("transit", transit.Addr),
					zap.String("src", src.Addr),
					zap.String("dst", dst),
					zap.String("call_path", cp.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method))
				rets.Release()
			}
		}()

	case callpath.Runtime:
		future, err := CallRuntime(p.svcCtx, cc, cp.Id, cp.Script, cp.Method, req.Args)
		if err != nil {
			log.L(p.svcCtx).Error("accept forwarded rpc notify to runtime failed",
				zap.String("transit", transit.Addr),
				zap.String("src", src.Addr),
				zap.String("dst", dst),
				zap.String("call_path", cp.String()),
				zap.String("id", cp.Id.String()),
				zap.String("script", cp.Script),
				zap.String("method", cp.Method),
				zap.Error(err))
			return
		}

		go func() {
			rets, err := waitAsyncResult(p.svcCtx, future)
			if err != nil {
				log.L(p.svcCtx).Error("accept forwarded rpc notify to runtime failed",
					zap.String("transit", transit.Addr),
					zap.String("src", src.Addr),
					zap.String("dst", dst),
					zap.String("call_path", cp.String()),
					zap.String("id", cp.Id.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method),
					zap.Error(err))
			} else {
				log.L(p.svcCtx).Debug("accept forwarded rpc notify to runtime finished",
					zap.String("transit", transit.Addr),
					zap.String("src", src.Addr),
					zap.String("dst", dst),
					zap.String("call_path", cp.String()),
					zap.String("id", cp.Id.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method))
				rets.Release()
			}
		}()

	case callpath.Entity:
		future, err := CallEntity(p.svcCtx, cc, cp.Id, cp.Script, cp.Method, req.Args)
		if err != nil {
			log.L(p.svcCtx).Error("accept forwarded rpc notify to entity failed",
				zap.String("transit", transit.Addr),
				zap.String("src", src.Addr),
				zap.String("dst", dst),
				zap.String("call_path", cp.String()),
				zap.String("id", cp.Id.String()),
				zap.String("script", cp.Script),
				zap.String("method", cp.Method),
				zap.Error(err))
			return
		}

		go func() {
			rets, err := waitAsyncResult(p.svcCtx, future)
			if err != nil {
				log.L(p.svcCtx).Error("accept forwarded rpc notify to entity failed",
					zap.String("transit", transit.Addr),
					zap.String("src", src.Addr),
					zap.String("dst", dst),
					zap.String("call_path", cp.String()),
					zap.String("id", cp.Id.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method),
					zap.Error(err))
			} else {
				log.L(p.svcCtx).Debug("accept forwarded rpc notify to entity finished",
					zap.String("transit", transit.Addr),
					zap.String("src", src.Addr),
					zap.String("dst", dst),
					zap.String("call_path", cp.String()),
					zap.String("id", cp.Id.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method))
				rets.Release()
			}
		}()
	}
}

func (p *_ForwardProcessor) acceptRequest(transit, src gap.Origin, dst string, req *gap.MsgRPCRequest) {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		err = fmt.Errorf("parse call path failed: %w", err)
		log.L(p.svcCtx).Error("accept forwarded rpc request failed",
			zap.String("transit", transit.Addr),
			zap.String("src", src.Addr),
			zap.String("dst", dst),
			zap.Int64("corr_id", req.CorrId),
			zap.Error(err))
		p.reply(transit, src, req.CorrId, nil, err)
		return
	}
	cp.Id = uid.From(dst)

	cc := rpcstack.CallChain{
		{
			Svc:       src.Svc,
			Addr:      src.Addr,
			Timestamp: time.UnixMilli(src.Timestamp).Local(),
			Transit:   false,
		},
		{
			Svc:       transit.Svc,
			Addr:      transit.Addr,
			Timestamp: time.UnixMilli(transit.Timestamp).Local(),
			Transit:   true,
		},
	}

	if len(p.permValidator) > 0 {
		passed, err := p.permValidator.SafeCall(func(passed bool, err error) bool {
			return !passed || err != nil
		}, cc, cp)
		if err != nil {
			err = fmt.Errorf("%w: %w", ErrPermissionDenied, err)
		} else if !passed {
			err = ErrPermissionDenied
		}
		if err != nil {
			err = fmt.Errorf("permission verification failed: %w", err)
			log.L(p.svcCtx).Error("accept forwarded rpc request failed",
				zap.String("transit", transit.Addr),
				zap.String("src", src.Addr),
				zap.String("dst", dst),
				zap.Int64("corr_id", req.CorrId),
				zap.String("call_path", cp.String()),
				zap.Error(err))
			p.reply(transit, src, req.CorrId, nil, err)
			return
		}
	}

	switch cp.TargetKind {
	case callpath.Service:
		go func() {
			rets, err := CallService(p.svcCtx, cc, cp.Script, cp.Method, req.Args)
			if err != nil {
				log.L(p.svcCtx).Error("accept forwarded rpc request to service failed",
					zap.String("transit", transit.Addr),
					zap.String("src", src.Addr),
					zap.String("dst", dst),
					zap.Int64("corr_id", req.CorrId),
					zap.String("call_path", cp.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method),
					zap.Error(err))
			} else {
				log.L(p.svcCtx).Debug("accept forwarded rpc request to service finished",
					zap.String("transit", transit.Addr),
					zap.String("src", src.Addr),
					zap.String("dst", dst),
					zap.Int64("corr_id", req.CorrId),
					zap.String("call_path", cp.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method))
			}
			p.reply(transit, src, req.CorrId, rets, err)
		}()

	case callpath.Runtime:
		future, err := CallRuntime(p.svcCtx, cc, cp.Id, cp.Script, cp.Method, req.Args)
		if err != nil {
			log.L(p.svcCtx).Error("accept forwarded rpc request to runtime failed",
				zap.String("transit", transit.Addr),
				zap.String("src", src.Addr),
				zap.String("dst", dst),
				zap.Int64("corr_id", req.CorrId),
				zap.String("call_path", cp.String()),
				zap.String("id", cp.Id.String()),
				zap.String("script", cp.Script),
				zap.String("method", cp.Method),
				zap.Error(err))
			p.reply(transit, src, req.CorrId, nil, err)
			return
		}

		go func() {
			rets, err := waitAsyncResult(p.svcCtx, future)
			if err != nil {
				log.L(p.svcCtx).Error("accept forwarded rpc request to runtime failed",
					zap.String("transit", transit.Addr),
					zap.String("src", src.Addr),
					zap.String("dst", dst),
					zap.Int64("corr_id", req.CorrId),
					zap.String("call_path", cp.String()),
					zap.String("id", cp.Id.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method),
					zap.Error(err))
			} else {
				log.L(p.svcCtx).Debug("accept forwarded rpc request to runtime finished",
					zap.String("transit", transit.Addr),
					zap.String("src", src.Addr),
					zap.String("dst", dst),
					zap.Int64("corr_id", req.CorrId),
					zap.String("call_path", cp.String()),
					zap.String("id", cp.Id.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method))
			}
			p.reply(transit, src, req.CorrId, rets, err)
		}()

	case callpath.Entity:
		future, err := CallEntity(p.svcCtx, cc, cp.Id, cp.Script, cp.Method, req.Args)
		if err != nil {
			log.L(p.svcCtx).Error("accept forwarded rpc request to entity failed",
				zap.String("transit", transit.Addr),
				zap.String("src", src.Addr),
				zap.String("dst", dst),
				zap.Int64("corr_id", req.CorrId),
				zap.String("call_path", cp.String()),
				zap.String("id", cp.Id.String()),
				zap.String("script", cp.Script),
				zap.String("method", cp.Method),
				zap.Error(err))
			p.reply(transit, src, req.CorrId, nil, err)
			return
		}

		go func() {
			rets, err := waitAsyncResult(p.svcCtx, future)
			if err != nil {
				log.L(p.svcCtx).Error("accept forwarded rpc request to entity failed",
					zap.String("transit", transit.Addr),
					zap.String("src", src.Addr),
					zap.String("dst", dst),
					zap.Int64("corr_id", req.CorrId),
					zap.String("call_path", cp.String()),
					zap.String("id", cp.Id.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method),
					zap.Error(err))
			} else {
				log.L(p.svcCtx).Debug("accept forwarded rpc request to entity finished",
					zap.String("transit", transit.Addr),
					zap.String("src", src.Addr),
					zap.String("dst", dst),
					zap.Int64("corr_id", req.CorrId),
					zap.String("call_path", cp.String()),
					zap.String("id", cp.Id.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method))
			}
			p.reply(transit, src, req.CorrId, rets, err)
		}()
	}
}

func (p *_ForwardProcessor) resolveReply(transit, src gap.Origin, reply *gap.MsgRPCReply) {
	ret := async.Result{}

	if reply.Error.OK() {
		if len(reply.Rets) > 0 {
			ret.Value = reply.Rets
		}
	} else {
		ret.Error = &reply.Error
	}

	if err := p.dsvc.FutureController().Resolve(reply.CorrId, ret); err != nil {
		log.L(p.svcCtx).Error("resolve forwarded rpc reply failed",
			zap.String("transit", transit.Addr),
			zap.String("src", src.Addr),
			zap.Int64("corr_id", reply.CorrId),
			zap.Error(err))
		return
	}

	log.L(p.svcCtx).Debug("forwarded rpc reply resolved",
		zap.String("transit", transit.Addr),
		zap.String("src", src.Addr),
		zap.Int64("corr_id", reply.CorrId))
}

func (p *_ForwardProcessor) reply(transit, src gap.Origin, corrId int64, rets variant.Array, retErr error) {
	defer rets.Release()

	if corrId == 0 {
		return
	}

	msg := &gap.MsgRPCReply{
		CorrId: corrId,
		Rets:   rets,
	}

	if retErr != nil {
		msg.Error = *variant.NewError(retErr)
	}

	msgBuf, err := gap.Marshal(msg)
	if err != nil {
		log.L(p.svcCtx).Error("marshal rpc reply failed",
			zap.String("transit", transit.Addr),
			zap.String("src", src.Addr),
			zap.Int64("corr_id", corrId),
			zap.Error(err))
		return
	}
	defer msgBuf.Release()

	forwardMsg := &gap.MsgForward{
		Dst:       src.Addr,
		CorrId:    corrId,
		TransId:   msg.MsgId(),
		TransData: msgBuf.Payload(),
	}

	if err := p.dsvc.Send(transit.Addr, forwardMsg); err != nil {
		log.L(p.svcCtx).Error("forward rpc reply failed",
			zap.String("transit", transit.Addr),
			zap.String("src", src.Addr),
			zap.Int64("corr_id", corrId),
			zap.Error(err))
		return
	}

	log.L(p.svcCtx).Debug("rpc reply forwarded",
		zap.String("src", src.Addr),
		zap.Int64("corr_id", corrId))
}
