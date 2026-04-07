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

package rpcli

import (
	"errors"
	"sync"

	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/addins/gate/cli"
	"git.golaxy.org/framework/addins/rpc/callpath"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/codec"
	"git.golaxy.org/framework/net/gap/variant"
	"go.uber.org/zap"
)

var (
	ErrScriptNotFound               = errors.New("rpcli: script not found")                // 找不到脚本
	ErrMethodNotFound               = errors.New("rpcli: method not found")                // 找不到方法
	ErrMethodParameterCountMismatch = errors.New("rpcli: method parameter count mismatch") // 方法参数数量不匹配
	ErrMethodParameterTypeMismatch  = errors.New("rpcli: method parameter type mismatch")  // 方法参数类型不匹配
)

// RPCli RCP客户端
type RPCli struct {
	*cli.Client
	encoder        *codec.Encoder
	decoder        *codec.Decoder
	remoteTime     cli.ResponseTime
	reduceCallPath bool
	scriptsMu      sync.RWMutex
	scripts        generic.SliceMap[string, IScript]
}

// RemoteTime 获取对端时间
func (c *RPCli) RemoteTime() cli.ResponseTime {
	return c.remoteTime
}

// RPC RPC调用
func (c *RPCli) RPC(service, comp, method string, args ...any) async.Future {
	handle, err := c.FutureController().New()
	if err != nil {
		return async.Return(async.NewFutureChan(), async.NewResult(nil, err))
	}

	vargs, err := variant.NewArray(args)
	if err != nil {
		handle.Cancel(err)
		return handle.Future()
	}

	cp := callpath.CallPath{
		TargetKind: callpath.Entity,
		Script:     comp,
		Method:     method,
	}

	cpBuf, err := cp.Encode(c.reduceCallPath)
	if err != nil {
		handle.Cancel(err)
		return handle.Future()
	}

	msg := &gap.MsgRPCRequest{
		CorrId: handle.Id(),
		Path:   cpBuf,
		Args:   vargs,
	}

	msgBuf, err := gap.Marshal(msg)
	if err != nil {
		handle.Cancel(err)
		return handle.Future()
	}
	defer msgBuf.Release()

	forwardMsg := &gap.MsgForward{
		Dst:       service,
		CorrId:    msg.CorrId,
		TransId:   msg.MsgId(),
		TransData: msgBuf.Payload(),
	}

	mpBuf, err := c.encoder.Encode(gap.Origin{Timestamp: c.remoteTime.NowTime().UnixMilli()}, 0, forwardMsg)
	if err != nil {
		handle.Cancel(err)
		return handle.Future()
	}
	defer mpBuf.Release()

	if err := c.DataIO().Send(mpBuf.Payload()); err != nil {
		handle.Cancel(err)
		return handle.Future()
	}

	c.L().Debug("rpc sent",
		zap.String("session_id", c.SessionId().String()),
		zap.String("local", c.NetAddr().Local.String()),
		zap.String("remote", c.NetAddr().Remote.String()),
		zap.String("dst", service),
		zap.Int64("corr_id", handle.Id()),
		zap.String("call_path", cp.String()))
	return handle.Future()
}

// OnewayRPC 单向RPC调用
func (c *RPCli) OnewayRPC(service, comp, method string, args ...any) error {
	vargs, err := variant.NewArray(args)
	if err != nil {
		return err
	}

	cp := callpath.CallPath{
		TargetKind: callpath.Entity,
		Script:     comp,
		Method:     method,
	}

	cpBuf, err := cp.Encode(c.reduceCallPath)
	if err != nil {
		return err
	}

	msg := &gap.MsgOnewayRPC{
		Path: cpBuf,
		Args: vargs,
	}

	msgBuf, err := gap.Marshal(msg)
	if err != nil {
		return err
	}
	defer msgBuf.Release()

	forwardMsg := &gap.MsgForward{
		Dst:       service,
		TransId:   msg.MsgId(),
		TransData: msgBuf.Payload(),
	}

	mpBuf, err := c.encoder.Encode(gap.Origin{Timestamp: c.remoteTime.NowTime().UnixMilli()}, 0, forwardMsg)
	if err != nil {
		return err
	}
	defer mpBuf.Release()

	if err := c.DataIO().Send(mpBuf.Payload()); err != nil {
		return err
	}

	c.L().Debug("oneway rpc sent",
		zap.String("session_id", c.SessionId().String()),
		zap.String("local", c.NetAddr().Local.String()),
		zap.String("remote", c.NetAddr().Remote.String()),
		zap.String("dst", service),
		zap.String("call_path", cp.String()))
	return nil
}
