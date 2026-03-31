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

	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/addins/dent"
	"git.golaxy.org/framework/addins/gate"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/addins/rpc/callpath"
	"git.golaxy.org/framework/addins/rpcstack"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/variant"
	"go.uber.org/zap"
)

// Match 是否匹配
func (p *_ForwardProcessor) Match(svcCtx service.Context, dst string, cc rpcstack.CallChain, cp callpath.CallPath, oneway bool) bool {
	// 只支持客户端域通信
	if !gate.ClientDetails.DomainRoot.Contains(dst) {
		return false
	}

	if oneway {
		// 单向请求，支持组播、单播地址
		return gate.ClientDetails.DomainUnicast.Contains(dst) || gate.ClientDetails.DomainMulticast.Contains(dst) || gate.ClientDetails.DomainBroadcast.Contains(dst)
	} else {
		// 普通请求，支持单播地址
		return gate.ClientDetails.DomainUnicast.Contains(dst)
	}
}

// Request 请求
func (p *_ForwardProcessor) Request(svcCtx service.Context, dst string, cc rpcstack.CallChain, cp callpath.CallPath, args []any) async.Future {
	handle, err := p.dsvc.FutureController().New()
	if err != nil {
		return async.Return(async.NewFutureChan(), async.NewResult(nil, err))
	}

	entityId, _ := gate.ClientDetails.DomainUnicast.Relative(dst)
	forwardAddr, err := p.getDistEntityForwardAddr(uid.From(entityId))
	if err != nil {
		handle.Cancel(err)
		return handle.Future()
	}

	vargs, err := variant.NewArray(args)
	if err != nil {
		handle.Cancel(err)
		return handle.Future()
	}

	cpBuf, err := cp.Encode(p.reduceCallPath)
	if err != nil {
		handle.Cancel(err)
		return handle.Future()
	}

	nextCC := append(cc, rpcstack.Call{
		Svc:       svcCtx.Name(),
		Addr:      p.dsvc.NodeDetails().LocalAddr,
		Timestamp: time.Now(),
		Transit:   false,
	})

	msg := &gap.MsgRPCRequest{
		CorrId:    handle.Id(),
		CallChain: nextCC,
		Path:      cpBuf,
		Args:      vargs,
	}

	msgBuf, err := gap.Marshal(msg)
	if err != nil {
		handle.Cancel(err)
		return handle.Future()
	}
	defer msgBuf.Release()

	forwardMsg := &gap.MsgForward{
		Dst:       dst,
		CorrId:    msg.CorrId,
		TransId:   msg.MsgId(),
		TransData: msgBuf.Payload(),
	}

	if err := p.dsvc.Send(forwardAddr, forwardMsg); err != nil {
		handle.Cancel(err)
		return handle.Future()
	}

	log.L(p.svcCtx).Debug("rpc request forwarded",
		zap.String("dst", dst),
		zap.Int64("corr_id", handle.Id()),
		zap.String("call_path", cp.String()))
	return handle.Future()
}

// Notify 通知
func (p *_ForwardProcessor) Notify(svcCtx service.Context, dst string, cc rpcstack.CallChain, cp callpath.CallPath, args []any) error {
	forwardAddr, err := p.getForwardAddr(dst)
	if err != nil {
		return err
	}

	vargs, err := variant.NewArray(args)
	if err != nil {
		return err
	}

	cpBuf, err := cp.Encode(p.reduceCallPath)
	if err != nil {
		return err
	}

	nextCC := append(cc, rpcstack.Call{
		Svc:       svcCtx.Name(),
		Addr:      p.dsvc.NodeDetails().LocalAddr,
		Timestamp: time.Now(),
		Transit:   false,
	})

	msg := &gap.MsgOnewayRPC{
		CallChain: nextCC,
		Path:      cpBuf,
		Args:      vargs,
	}

	msgBuf, err := gap.Marshal(msg)
	if err != nil {
		return err
	}
	defer msgBuf.Release()

	forwardMsg := &gap.MsgForward{
		Dst:       dst,
		TransId:   msg.MsgId(),
		TransData: msgBuf.Payload(),
	}

	if err := p.dsvc.Send(forwardAddr, forwardMsg); err != nil {
		return err
	}

	log.L(p.svcCtx).Debug("rpc notify forwarded",
		zap.String("dst", dst),
		zap.String("call_path", cp.String()))
	return nil
}

func (p *_ForwardProcessor) getForwardAddr(dst string) (string, error) {
	nodeId, ok := gate.ClientDetails.DomainUnicast.Relative(dst)
	if ok {
		// 目标为单播地址，查询实体的通信中转服务地址
		return p.getDistEntityForwardAddr(uid.From(nodeId))
	}

	if gate.ClientDetails.DomainMulticast.Contains(dst) || gate.ClientDetails.DomainBroadcast.Contains(dst) {
		// 目标为组播地址，广播所有的通信中转服务
		return p.transitBroadcastAddr, nil
	}

	return "", ErrIncorrectDestAddress
}

func (p *_ForwardProcessor) getDistEntityForwardAddr(entId uid.Id) (string, error) {
	distEntity, ok := p.dentq.GetDistEntity(entId)
	if !ok {
		return "", ErrDistEntityNotFound
	}

	idx := slices.IndexFunc(distEntity.Nodes, func(node dent.Node) bool {
		return node.Service == p.transitService
	})
	if idx < 0 {
		return "", ErrDistEntityNodeNotFound
	}

	return distEntity.Nodes[idx].RemoteAddr, nil
}
