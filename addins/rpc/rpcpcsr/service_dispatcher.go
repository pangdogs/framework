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
	"context"
	"fmt"
	"time"

	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/addins/rpc/callpath"
	"git.golaxy.org/framework/addins/rpcstack"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/variant"
	"git.golaxy.org/framework/utils/correlation"
	"go.uber.org/zap"
)

// handleServiceMsg 过滤非服务域来源，并分发请求、通知和响应。
func (p *_ServiceProcessor) handleServiceMsg(topic string, mp gap.MsgPacket) {
	// 只支持服务域通信
	if !p.dsvc.NodeDetails().DomainRoot.Contains(mp.Head.Src.Addr) {
		return
	}

	switch mp.Head.MsgID {
	case gap.MsgID_OnewayRPC:
		p.acceptNotify(mp.Head.Src, mp.Body.(*gap.MsgOnewayRPC))

	case gap.MsgID_RPC_Request:
		p.acceptRequest(mp.Head.Src, mp.Body.(*gap.MsgRPCRequest))

	case gap.MsgID_RPC_Reply:
		p.resolveReply(mp.Head.Src, mp.Body.(*gap.MsgRPCReply))
	}
}

// acceptNotify 校验来源、排除规则和权限后，异步调用服务、运行时或实体目标。
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
		spawnProcessorTask(p.svcCtx, p.scope, func(context.Context) {
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
			}
			rets.ReleaseIfSnapshot()
		})

	case callpath.Runtime:
		future, err := CallRuntime(p.svcCtx, cc, cp.ID, cp.Script, cp.Method, req.Args)
		if err != nil {
			log.L(p.svcCtx).Error("accept rpc notify to runtime failed",
				zap.String("src", src.Addr),
				zap.String("call_path", cp.String()),
				zap.String("id", cp.ID.String()),
				zap.String("script", cp.Script),
				zap.String("method", cp.Method),
				zap.Error(err))
			return
		}

		spawnProcessorTask(p.svcCtx, p.scope, func(ctx context.Context) {
			rets, err := waitAsyncResult(ctx, future)
			if err != nil {
				log.L(p.svcCtx).Error("accept rpc notify to runtime failed",
					zap.String("src", src.Addr),
					zap.String("call_path", cp.String()),
					zap.String("id", cp.ID.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method),
					zap.Error(err))
			} else {
				log.L(p.svcCtx).Debug("accept rpc notify to runtime finished",
					zap.String("src", src.Addr),
					zap.String("call_path", cp.String()),
					zap.String("id", cp.ID.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method))
			}
			rets.ReleaseIfSnapshot()
		})

	case callpath.Entity:
		future, err := CallEntity(p.svcCtx, cc, cp.ID, cp.Script, cp.Method, req.Args)
		if err != nil {
			log.L(p.svcCtx).Error("accept rpc notify to entity failed",
				zap.String("src", src.Addr),
				zap.String("call_path", cp.String()),
				zap.String("id", cp.ID.String()),
				zap.String("script", cp.Script),
				zap.String("method", cp.Method),
				zap.Error(err))
			return
		}

		spawnProcessorTask(p.svcCtx, p.scope, func(ctx context.Context) {
			rets, err := waitAsyncResult(ctx, future)
			if err != nil {
				log.L(p.svcCtx).Error("accept rpc notify to entity failed",
					zap.String("src", src.Addr),
					zap.String("call_path", cp.String()),
					zap.String("id", cp.ID.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method),
					zap.Error(err))
			} else {
				log.L(p.svcCtx).Debug("accept rpc notify to entity finished",
					zap.String("src", src.Addr),
					zap.String("call_path", cp.String()),
					zap.String("id", cp.ID.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method))
			}
			rets.ReleaseIfSnapshot()
		})
	}
}

// acceptRequest 校验权限并异步调用目标，完成后向来源地址发送响应。
func (p *_ServiceProcessor) acceptRequest(src gap.Origin, req *gap.MsgRPCRequest) {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		err = fmt.Errorf("parse call path failed: %w", err)
		log.L(p.svcCtx).Error("accept rpc request failed",
			zap.String("src", src.Addr),
			zap.Uint64("corr_id", uint64(req.CorrID)),
			zap.Error(err))
		p.reply(src, req.CorrID, variant.Array{}, err)
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
				zap.Uint64("corr_id", uint64(req.CorrID)),
				zap.String("call_path", cp.String()),
				zap.Error(err))
			p.reply(src, req.CorrID, variant.Array{}, err)
			return
		}
	}

	switch cp.TargetKind {
	case callpath.Service:
		spawnProcessorTask(p.svcCtx, p.scope, func(context.Context) {
			rets, err := CallService(p.svcCtx, cc, cp.Script, cp.Method, req.Args)
			if err != nil {
				log.L(p.svcCtx).Error("accept rpc request to service failed",
					zap.String("src", src.Addr),
					zap.Uint64("corr_id", uint64(req.CorrID)),
					zap.String("call_path", cp.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method),
					zap.Error(err))
			} else {
				log.L(p.svcCtx).Debug("accept rpc request to service finished",
					zap.String("src", src.Addr),
					zap.Uint64("corr_id", uint64(req.CorrID)),
					zap.String("call_path", cp.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method))
			}
			p.reply(src, req.CorrID, rets, err)
		})

	case callpath.Runtime:
		future, err := CallRuntime(p.svcCtx, cc, cp.ID, cp.Script, cp.Method, req.Args)
		if err != nil {
			log.L(p.svcCtx).Error("accept rpc request to runtime failed",
				zap.String("src", src.Addr),
				zap.Uint64("corr_id", uint64(req.CorrID)),
				zap.String("call_path", cp.String()),
				zap.String("id", cp.ID.String()),
				zap.String("script", cp.Script),
				zap.String("method", cp.Method),
				zap.Error(err))
			p.reply(src, req.CorrID, variant.Array{}, err)
			return
		}

		spawnProcessorTask(p.svcCtx, p.scope, func(ctx context.Context) {
			rets, err := waitAsyncResult(ctx, future)
			if err != nil {
				log.L(p.svcCtx).Error("accept rpc request to runtime failed",
					zap.String("src", src.Addr),
					zap.Uint64("corr_id", uint64(req.CorrID)),
					zap.String("call_path", cp.String()),
					zap.String("id", cp.ID.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method),
					zap.Error(err))
			} else {
				log.L(p.svcCtx).Debug("accept rpc request to runtime finished",
					zap.String("src", src.Addr),
					zap.Uint64("corr_id", uint64(req.CorrID)),
					zap.String("call_path", cp.String()),
					zap.String("id", cp.ID.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method))
			}
			p.reply(src, req.CorrID, rets, err)
		})

	case callpath.Entity:
		future, err := CallEntity(p.svcCtx, cc, cp.ID, cp.Script, cp.Method, req.Args)
		if err != nil {
			log.L(p.svcCtx).Error("accept rpc request to entity failed",
				zap.String("src", src.Addr),
				zap.Uint64("corr_id", uint64(req.CorrID)),
				zap.String("call_path", cp.String()),
				zap.String("id", cp.ID.String()),
				zap.String("script", cp.Script),
				zap.String("method", cp.Method),
				zap.Error(err))
			p.reply(src, req.CorrID, variant.Array{}, err)
			return
		}

		spawnProcessorTask(p.svcCtx, p.scope, func(ctx context.Context) {
			rets, err := waitAsyncResult(ctx, future)
			if err != nil {
				log.L(p.svcCtx).Error("accept rpc request to entity failed",
					zap.String("src", src.Addr),
					zap.Uint64("corr_id", uint64(req.CorrID)),
					zap.String("call_path", cp.String()),
					zap.String("id", cp.ID.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method),
					zap.Error(err))
			} else {
				log.L(p.svcCtx).Debug("accept rpc request to entity finished",
					zap.String("src", src.Addr),
					zap.Uint64("corr_id", uint64(req.CorrID)),
					zap.String("call_path", cp.String()),
					zap.String("id", cp.ID.String()),
					zap.String("script", cp.Script),
					zap.String("method", cp.Method))
			}
			p.reply(src, req.CorrID, rets, err)
		})
	}
}

// resolveReply 按关联 ID 完成本节点发起服务 RPC 时创建的 Future。
func (p *_ServiceProcessor) resolveReply(src gap.Origin, reply *gap.MsgRPCReply) {
	ret := async.Result{}

	if reply.Error.OK() {
		if len(reply.Rets.Items) > 0 {
			ret.Value = reply.Rets
		}
	} else {
		ret.Error = &reply.Error
	}

	if !p.dsvc.Correlation().Resolve(reply.CorrID, ret) {
		log.L(p.svcCtx).Error("resolve rpc reply failed",
			zap.String("src", src.Addr),
			zap.Uint64("corr_id", uint64(reply.CorrID)))
		return
	}

	log.L(p.svcCtx).Debug("rpc reply resolved",
		zap.String("src", src.Addr),
		zap.Uint64("corr_id", uint64(reply.CorrID)))
}

// reply 向请求来源发送结果并释放临时返回值快照；零关联 ID 不回复。
func (p *_ServiceProcessor) reply(src gap.Origin, corrID correlation.ID, rets variant.Array, retErr error) {
	defer rets.ReleaseIfSnapshot()

	if corrID == 0 {
		return
	}

	msg := &gap.MsgRPCReply{
		CorrID: corrID,
		Rets:   rets,
	}

	if retErr != nil {
		msg.Error = *variant.NewError(retErr)
	}

	if err := p.dsvc.Send(src.Addr, msg); err != nil {
		log.L(p.svcCtx).Error("rpc reply failed",
			zap.String("src", src.Addr),
			zap.Uint64("corr_id", uint64(corrID)),
			zap.Error(err))
		return
	}

	log.L(p.svcCtx).Debug("rpc reply sent",
		zap.String("src", src.Addr),
		zap.Uint64("corr_id", uint64(corrID)))
}
