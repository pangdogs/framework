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
	// ErrScriptNotFound 表示找不到目标客户端脚本。
	ErrScriptNotFound = errors.New("rpcli: script not found")
	// ErrMethodNotFound 表示找不到目标脚本方法。
	ErrMethodNotFound = errors.New("rpcli: method not found")
	// ErrMethodParameterCountMismatch 表示调用参数数量与方法签名不匹配。
	ErrMethodParameterCountMismatch = errors.New("rpcli: method parameter count mismatch")
	// ErrMethodParameterTypeMismatch 表示调用参数类型与方法签名不匹配。
	ErrMethodParameterTypeMismatch = errors.New("rpcli: method parameter type mismatch")
)

// RPCli 是通过网关连接服务的 RPC 客户端。
type RPCli struct {
	*cli.Client
	encoder        *codec.Encoder
	decoder        *codec.Decoder
	remoteClock    cli.TimeSample
	reduceCallPath bool
	scriptsMu      sync.RWMutex
	scripts        generic.SliceMap[string, IScript]
}

// RemoteClock 返回连接建立时选出的最低 RTT 对端时钟样本。
func (c *RPCli) RemoteClock() cli.TimeSample {
	return c.remoteClock
}

// RPC 向服务的实体目标发起请求，并返回用于接收响应的 Future。
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

	mpBuf, err := c.encoder.Encode(gap.Origin{Timestamp: c.remoteClock.RemoteNow().UnixMilli()}, 0, forwardMsg)
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

// OnewayRPC 向服务的实体目标发送无需响应的通知。
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

	mpBuf, err := c.encoder.Encode(gap.Origin{Timestamp: c.remoteClock.RemoteNow().UnixMilli()}, 0, forwardMsg)
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
