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
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/variant"
	"git.golaxy.org/framework/plugins/dentq"
	"git.golaxy.org/framework/plugins/gate"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/plugins/rpcstack"
	"git.golaxy.org/framework/utils/concurrent"
	"slices"
)

// Match 是否匹配
func (p *_ForwardProcessor) Match(ctx service.Context, dst string, callChain rpcstack.CallChain, path string, oneWay bool) bool {
	// 只支持客户端域通信
	if !gate.CliDetails.DomainRoot.Contains(dst) {
		return false
	}

	if oneWay {
		// 单向请求，支持组播、单播地址
		return gate.CliDetails.DomainUnicast.Contains(dst) || gate.CliDetails.DomainMulticast.Contains(dst) || gate.CliDetails.DomainBroadcast.Contains(dst)
	} else {
		// 普通请求，支持单播地址
		return gate.CliDetails.DomainUnicast.Contains(dst)
	}
}

// Request 请求
func (p *_ForwardProcessor) Request(ctx service.Context, dst string, callChain rpcstack.CallChain, path string, args []any) async.AsyncRet {
	ret := concurrent.MakeRespAsyncRet()
	future := concurrent.MakeFuture(p.dist.GetFutures(), nil, ret)

	entId, _ := gate.CliDetails.DomainUnicast.Relative(dst)
	forwardAddr, err := p.getDistEntityForwardAddr(uid.From(entId))
	if err != nil {
		future.Cancel(err)
		return ret.ToAsyncRet()
	}

	vargs, err := variant.MakeReadonlyArray(args)
	if err != nil {
		future.Cancel(err)
		return ret.ToAsyncRet()
	}

	msg := &gap.MsgRPCRequest{
		CorrId:    future.Id,
		CallChain: callChain,
		Path:      path,
		Args:      vargs,
	}

	bs, err := gap.Marshal(msg)
	if err != nil {
		future.Cancel(err)
		return ret.ToAsyncRet()
	}
	defer bs.Release()

	forwardMsg := &gap.MsgForward{
		Dst:       dst,
		CorrId:    msg.CorrId,
		TransId:   msg.MsgId(),
		TransData: bs.Data(),
	}

	if err = p.dist.SendMsg(forwardAddr, forwardMsg); err != nil {
		future.Cancel(err)
		return ret.ToAsyncRet()
	}

	log.Debugf(p.servCtx, "rpc request(%d) forwarding to dst:%q, path:%q ok", future.Id, forwardAddr, path)
	return ret.ToAsyncRet()
}

// Notify 通知
func (p *_ForwardProcessor) Notify(ctx service.Context, dst string, callChain rpcstack.CallChain, path string, args []any) error {
	forwardAddr, err := p.getForwardAddr(dst)
	if err != nil {
		return err
	}

	vargs, err := variant.MakeReadonlyArray(args)
	if err != nil {
		return err
	}

	msg := &gap.MsgOnewayRPC{
		CallChain: callChain,
		Path:      path,
		Args:      vargs,
	}

	bs, err := gap.Marshal(msg)
	if err != nil {
		return err
	}
	defer bs.Release()

	forwardMsg := &gap.MsgForward{
		Dst:       dst,
		TransId:   msg.MsgId(),
		TransData: bs.Data(),
	}

	if err = p.dist.SendMsg(forwardAddr, forwardMsg); err != nil {
		return err
	}

	log.Debugf(p.servCtx, "rpc notify forwarding to dst:%q, path:%q ok", forwardAddr, path)
	return nil
}

func (p *_ForwardProcessor) getForwardAddr(dst string) (string, error) {
	nodeId, ok := gate.CliDetails.DomainUnicast.Relative(dst)
	if ok {
		// 目标为单播地址，查询实体的通信中转服务地址
		return p.getDistEntityForwardAddr(uid.From(nodeId))
	}

	if gate.CliDetails.DomainMulticast.Contains(dst) || gate.CliDetails.DomainBroadcast.Contains(dst) {
		// 目标为组播地址，广播所有的通信中转服务
		return p.transitBroadcastAddr, nil

	} else {
		return "", ErrIncorrectDestAddress
	}
}

func (p *_ForwardProcessor) getDistEntityForwardAddr(entId uid.Id) (string, error) {
	dent, ok := p.dentq.GetDistEntity(entId)
	if !ok {
		return "", ErrDistEntityNotFound
	}

	idx := slices.IndexFunc(dent.Nodes, func(node dentq.Node) bool {
		return node.Service == p.transitService
	})
	if idx < 0 {
		return "", ErrDistEntityNodeNotFound
	}

	return dent.Nodes[idx].RemoteAddr, nil
}
