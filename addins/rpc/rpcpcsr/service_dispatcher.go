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
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/addins/rpc/callpath"
	"git.golaxy.org/framework/addins/rpcstack"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/variant"
	"go.uber.org/zap"
)

func (p *_ServiceProcessor) handleServiceMsg(topic string, mp gap.MsgPacket) {
	// 只支持服务域通信
	if !p.dsvc.NodeDetails().DomainRoot.Contains(mp.Head.Src.Addr) {
		return
	}

	switch mp.Head.MsgId {
	case gap.MsgId_OnewayRPC:
		p.acceptNotify(mp.Head.Src, mp.Body.(*gap.MsgOnewayRPC))

	case gap.MsgId_RPC_Request:
		p.acceptRequest(mp.Head.Src, mp.Body.(*gap.MsgRPCRequest))

	case gap.MsgId_RPC_Reply:
		p.resolveReply(mp.Head.Src, mp.Body.(*gap.MsgRPCReply))
	}
}

func (p *_ServiceProcessor) acceptNotify(src gap.Origin, req *gap.MsgOnewayRPC) {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		log.L(p.svcCtx).Error("accept rpc notify failed",
			zap.String("src", src.Addr),
			zap.Error(fmt.Errorf("parse call path failed: %w", err)))
		return
	}

	if cp.ExcludeSrc && src.Addr == p.dsvc.NodeDetails().LocalAddr {
		log.L(p.svcCtx).Debug("accept rpc notify skipped, source excluded",
			zap.String("src", src.Addr),
			zap.String("call_path", cp.String()))
		return
	}

	cc := append(req.CallChain,
		rpcstack.Call{
			Svc:       src.Svc,
			Addr:      src.Addr,
			Timestamp: time.UnixMilli(src.Timestamp).Local(),
			Transit:   false,
		},
	)

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
			log.L(p.svcCtx).Error("accept rpc notify failed",
				zap.String("src", src.Addr),
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
				log.L(p.svcCtx).Error("accept rpc notify to service failed",
					zap.String("src", src.Addr),
					zap.String("call_path", cp.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method),
					zap.Error(err))
			} else {
				log.L(p.svcCtx).Debug("accept rpc notify to service finished",
					zap.String("src", src.Addr),
					zap.String("call_path", cp.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method))
				rets.Release()
			}
		}()

	case callpath.Runtime:
		future, err := CallRuntime(p.svcCtx, cc, cp.Id, cp.Script, cp.Method, req.Args)
		if err != nil {
			log.L(p.svcCtx).Error("accept rpc notify to runtime failed",
				zap.String("src", src.Addr),
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
				log.L(p.svcCtx).Error("accept rpc notify to runtime failed",
					zap.String("src", src.Addr),
					zap.String("call_path", cp.String()),
					zap.String("id", cp.Id.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method),
					zap.Error(err))
			} else {
				log.L(p.svcCtx).Debug("accept rpc notify to runtime finished",
					zap.String("src", src.Addr),
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
			log.L(p.svcCtx).Error("accept rpc notify to entity failed",
				zap.String("src", src.Addr),
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
				log.L(p.svcCtx).Error("accept rpc notify to entity failed",
					zap.String("src", src.Addr),
					zap.String("call_path", cp.String()),
					zap.String("id", cp.Id.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method),
					zap.Error(err))
			} else {
				log.L(p.svcCtx).Debug("accept rpc notify to entity finished",
					zap.String("src", src.Addr),
					zap.String("call_path", cp.String()),
					zap.String("id", cp.Id.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method))
				rets.Release()
			}
		}()
	}
}

func (p *_ServiceProcessor) acceptRequest(src gap.Origin, req *gap.MsgRPCRequest) {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		err = fmt.Errorf("parse call path failed: %w", err)
		log.L(p.svcCtx).Error("accept rpc request failed",
			zap.String("src", src.Addr),
			zap.Int64("corr_id", req.CorrId),
			zap.Error(err))
		p.reply(src, req.CorrId, variant.SerializedArray{}, err)
		return
	}

	cc := append(req.CallChain,
		rpcstack.Call{
			Svc:       src.Svc,
			Addr:      src.Addr,
			Timestamp: time.UnixMilli(src.Timestamp).Local(),
			Transit:   false,
		},
	)

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
			log.L(p.svcCtx).Error("accept rpc request failed",
				zap.String("src", src.Addr),
				zap.Int64("corr_id", req.CorrId),
				zap.String("call_path", cp.String()),
				zap.Error(err))
			p.reply(src, req.CorrId, variant.SerializedArray{}, err)
			return
		}
	}

	switch cp.TargetKind {
	case callpath.Service:
		go func() {
			rets, err := CallService(p.svcCtx, cc, cp.Script, cp.Method, req.Args)
			if err != nil {
				log.L(p.svcCtx).Error("accept rpc request to service failed",
					zap.String("src", src.Addr),
					zap.Int64("corr_id", req.CorrId),
					zap.String("call_path", cp.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method),
					zap.Error(err))
			} else {
				log.L(p.svcCtx).Debug("accept rpc request to service finished",
					zap.String("src", src.Addr),
					zap.Int64("corr_id", req.CorrId),
					zap.String("call_path", cp.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method))
			}
			p.reply(src, req.CorrId, rets, err)
		}()

	case callpath.Runtime:
		future, err := CallRuntime(p.svcCtx, cc, cp.Id, cp.Script, cp.Method, req.Args)
		if err != nil {
			log.L(p.svcCtx).Error("accept rpc request to runtime failed",
				zap.String("src", src.Addr),
				zap.Int64("corr_id", req.CorrId),
				zap.String("call_path", cp.String()),
				zap.String("id", cp.Id.String()),
				zap.String("script", cp.Script),
				zap.String("method", cp.Method),
				zap.Error(err))
			p.reply(src, req.CorrId, variant.SerializedArray{}, err)
			return
		}

		go func() {
			rets, err := waitAsyncResult(p.svcCtx, future)
			if err != nil {
				log.L(p.svcCtx).Error("accept rpc request to runtime failed",
					zap.String("src", src.Addr),
					zap.Int64("corr_id", req.CorrId),
					zap.String("call_path", cp.String()),
					zap.String("id", cp.Id.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method),
					zap.Error(err))
			} else {
				log.L(p.svcCtx).Debug("accept rpc request to runtime finished",
					zap.String("src", src.Addr),
					zap.Int64("corr_id", req.CorrId),
					zap.String("call_path", cp.String()),
					zap.String("id", cp.Id.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method))
			}
			p.reply(src, req.CorrId, rets, err)
		}()

	case callpath.Entity:
		future, err := CallEntity(p.svcCtx, cc, cp.Id, cp.Script, cp.Method, req.Args)
		if err != nil {
			log.L(p.svcCtx).Error("accept rpc request to entity failed",
				zap.String("src", src.Addr),
				zap.Int64("corr_id", req.CorrId),
				zap.String("call_path", cp.String()),
				zap.String("id", cp.Id.String()),
				zap.String("script", cp.Script),
				zap.String("method", cp.Method),
				zap.Error(err))
			p.reply(src, req.CorrId, variant.SerializedArray{}, err)
			return
		}

		go func() {
			rets, err := waitAsyncResult(p.svcCtx, future)
			if err != nil {
				log.L(p.svcCtx).Error("accept rpc request to entity failed",
					zap.String("src", src.Addr),
					zap.Int64("corr_id", req.CorrId),
					zap.String("call_path", cp.String()),
					zap.String("id", cp.Id.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method),
					zap.Error(err))
			} else {
				log.L(p.svcCtx).Debug("accept rpc request to entity finished",
					zap.String("src", src.Addr),
					zap.Int64("corr_id", req.CorrId),
					zap.String("call_path", cp.String()),
					zap.String("id", cp.Id.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method))
			}
			p.reply(src, req.CorrId, rets, err)
		}()
	}
}

func (p *_ServiceProcessor) resolveReply(src gap.Origin, reply *gap.MsgRPCReply) {
	ret := async.Result{}

	if reply.Error.OK() {
		if len(reply.Rets) > 0 {
			ret.Value = reply.Rets
		}
	} else {
		ret.Error = &reply.Error
	}

	if err := p.dsvc.FutureController().Resolve(reply.CorrId, ret); err != nil {
		log.L(p.svcCtx).Error("resolve rpc reply failed",
			zap.String("src", src.Addr),
			zap.Int64("corr_id", reply.CorrId),
			zap.Error(err))
		return
	}

	log.L(p.svcCtx).Debug("rpc reply resolved",
		zap.String("src", src.Addr),
		zap.Int64("corr_id", reply.CorrId))
}

func (p *_ServiceProcessor) reply(src gap.Origin, corrId int64, rets variant.SerializedArray, retErr error) {
	defer rets.Release()

	if corrId == 0 {
		return
	}

	msg := &gap.MsgRPCReply{
		CorrId: corrId,
		Rets:   rets.Ref(),
	}

	if retErr != nil {
		msg.Error = *variant.NewError(retErr)
	}

	if err := p.dsvc.Send(src.Addr, msg); err != nil {
		log.L(p.svcCtx).Error("rpc reply failed",
			zap.String("src", src.Addr),
			zap.Int64("corr_id", corrId),
			zap.Error(err))
		return
	}

	log.L(p.svcCtx).Debug("rpc reply sent",
		zap.String("src", src.Addr),
		zap.Int64("corr_id", corrId))
}
