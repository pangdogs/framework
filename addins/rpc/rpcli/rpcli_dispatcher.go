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
	"fmt"
	"reflect"
	"time"

	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/framework/addins/rpc/callpath"
	"git.golaxy.org/framework/addins/rpcstack"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/variant"
	"go.uber.org/zap"
)

var (
	callChainRT = reflect.TypeFor[rpcstack.CallChain]()
)

func (c *RPCli) handleData(data []byte) {
	mp, err := c.decoder.Decode(data)
	if err != nil {
		c.Logger().Error("decode data failed",
			zap.String("session_id", c.SessionId().String()),
			zap.Error(err))
		return
	}

	switch mp.Head.MsgId {
	case gap.MsgId_OnewayRPC:
		c.acceptNotify(mp.Head.Src, mp.Msg.(*gap.MsgOnewayRPC))

	case gap.MsgId_RPC_Request:
		c.acceptRequest(mp.Head.Src, mp.Msg.(*gap.MsgRPCRequest))

	case gap.MsgId_RPC_Reply:
		c.resolveReply(mp.Msg.(*gap.MsgRPCReply))
	}
}

func (c *RPCli) acceptNotify(src gap.Origin, req *gap.MsgOnewayRPC) {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		c.Logger().Error("accept rpc notify failed",
			zap.String("session_id", c.SessionId().String()),
			zap.String("local", c.NetAddr().Local.String()),
			zap.String("remote", c.NetAddr().Remote.String()),
			zap.Error(fmt.Errorf("parse call path failed: %w", err)))
		return
	}

	cc := append(req.CallChain,
		rpcstack.Call{
			Svc:       src.Svc,
			Addr:      src.Addr,
			Timestamp: time.UnixMilli(src.Timestamp).Local(),
			Transit:   true,
		},
	)

	switch cp.TargetKind {
	case callpath.Client:
		rets, err := c.callScript(cc, cp.Script, cp.Method, req.Args)
		if err != nil {
			c.Logger().Error("accept rpc notify failed",
				zap.String("session_id", c.SessionId().String()),
				zap.String("local", c.NetAddr().Local.String()),
				zap.String("remote", c.NetAddr().Remote.String()),
				zap.String("script", cp.Script),
				zap.String("method", cp.Method),
				zap.Error(err))
		} else {
			c.Logger().Debug("accept rpc notify finished",
				zap.String("session_id", c.SessionId().String()),
				zap.String("local", c.NetAddr().Local.String()),
				zap.String("remote", c.NetAddr().Remote.String()),
				zap.String("call_path", cp.String()),
				zap.String("script", cp.Script),
				zap.String("method", cp.Method))
			rets.Release()
		}
	}
}

func (c *RPCli) acceptRequest(src gap.Origin, req *gap.MsgRPCRequest) {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		err = fmt.Errorf("parse call path failed: %w", err)
		c.Logger().Error("accept rpc request failed",
			zap.String("session_id", c.SessionId().String()),
			zap.String("local", c.NetAddr().Local.String()),
			zap.String("remote", c.NetAddr().Remote.String()),
			zap.Int64("corr_id", req.CorrId),
			zap.Error(err))
		c.reply(src, req.CorrId, nil, err)
		return
	}

	cc := append(req.CallChain,
		rpcstack.Call{
			Svc:       src.Svc,
			Addr:      src.Addr,
			Timestamp: time.UnixMilli(src.Timestamp).Local(),
			Transit:   true,
		},
	)

	switch cp.TargetKind {
	case callpath.Client:
		rets, err := c.callScript(cc, cp.Script, cp.Method, req.Args)
		if err != nil {
			c.Logger().Error("accept rpc request failed",
				zap.String("session_id", c.SessionId().String()),
				zap.String("local", c.NetAddr().Local.String()),
				zap.String("remote", c.NetAddr().Remote.String()),
				zap.Int64("corr_id", req.CorrId),
				zap.String("call_path", cp.String()),
				zap.String("script", cp.Script),
				zap.String("method", cp.Method),
				zap.Error(err))
		} else {
			c.Logger().Debug("accept rpc request finished",
				zap.String("session_id", c.SessionId().String()),
				zap.String("local", c.NetAddr().Local.String()),
				zap.String("remote", c.NetAddr().Remote.String()),
				zap.Int64("corr_id", req.CorrId),
				zap.String("call_path", cp.String()),
				zap.String("script", cp.Script),
				zap.String("method", cp.Method))
		}
		c.reply(src, req.CorrId, rets, err)
	}
}

func (c *RPCli) resolveReply(reply *gap.MsgRPCReply) {
	ret := async.Result{}

	if reply.Error.OK() {
		if len(reply.Rets) > 0 {
			ret.Value = reply.Rets
		}
	} else {
		ret.Error = &reply.Error
	}

	if err := c.FutureController().Resolve(reply.CorrId, ret); err != nil {
		c.Logger().Error("resolve rpc reply failed",
			zap.String("session_id", c.SessionId().String()),
			zap.String("local", c.NetAddr().Local.String()),
			zap.String("remote", c.NetAddr().Remote.String()),
			zap.Int64("corr_id", reply.CorrId),
			zap.Error(err))
		return
	}

	c.Logger().Debug("rpc reply resolved",
		zap.String("session_id", c.SessionId().String()),
		zap.String("local", c.NetAddr().Local.String()),
		zap.String("remote", c.NetAddr().Remote.String()),
		zap.Int64("corr_id", reply.CorrId))
}

func (c *RPCli) reply(src gap.Origin, corrId int64, rets variant.Array, retErr error) {
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
		c.Logger().Error("marshal rpc reply failed",
			zap.String("session_id", c.SessionId().String()),
			zap.String("local", c.NetAddr().Local.String()),
			zap.String("remote", c.NetAddr().Remote.String()),
			zap.Int64("corr_id", corrId),
			zap.Error(err))
		return
	}
	defer msgBuf.Release()

	forwardMsg := &gap.MsgForward{
		Dst:       src.Addr,
		CorrId:    msg.CorrId,
		TransId:   msg.MsgId(),
		TransData: msgBuf.Payload(),
	}

	mpBuf, err := c.encoder.Encode(gap.Origin{Timestamp: c.remoteTime.NowTime().UnixMilli()}, 0, forwardMsg)
	if err != nil {
		c.Logger().Error("encode rpc reply failed",
			zap.String("session_id", c.SessionId().String()),
			zap.String("local", c.NetAddr().Local.String()),
			zap.String("remote", c.NetAddr().Remote.String()),
			zap.Int64("corr_id", corrId),
			zap.Error(err))
		return
	}
	defer mpBuf.Release()

	if err = c.DataIO().Send(mpBuf.Payload()); err != nil {
		c.Logger().Error("send rpc reply failed",
			zap.String("session_id", c.SessionId().String()),
			zap.String("local", c.NetAddr().Local.String()),
			zap.String("remote", c.NetAddr().Remote.String()),
			zap.Int64("corr_id", corrId),
			zap.Error(err))
		return
	}

	c.Logger().Debug("rpc reply sent",
		zap.String("session_id", c.SessionId().String()),
		zap.String("local", c.NetAddr().Local.String()),
		zap.String("remote", c.NetAddr().Remote.String()),
		zap.Int64("corr_id", corrId))
}

func (c *RPCli) callScript(cc rpcstack.CallChain, script, method string, args variant.Array) (rets variant.Array, err error) {
	scr, ok := c.scripts.Get(script)
	if !ok {
		return nil, ErrScriptNotFound
	}

	methodRV := scr.Reflected().MethodByName(method)
	if !methodRV.IsValid() {
		return nil, ErrMethodNotFound
	}

	argsRV, err := parseArgs(methodRV, cc, args)
	if err != nil {
		return nil, err
	}

	return variant.NewSerializedArray(methodRV.Call(argsRV))
}

func parseArgs(methodRV reflect.Value, cc rpcstack.CallChain, args variant.Array) ([]reflect.Value, error) {
	methodRT := methodRV.Type()
	ccPos := -1

	for i := range methodRT.NumIn() {
		if !callChainRT.AssignableTo(methodRT.In(i)) {
			continue
		}
		if ccPos >= 0 {
			return nil, ErrMethodParameterCountMismatch
		}
		ccPos = i
	}

	switch {
	case ccPos < 0 && methodRT.NumIn() != len(args):
		return nil, ErrMethodParameterCountMismatch
	case ccPos >= 0 && methodRT.NumIn() != len(args)+1:
		return nil, ErrMethodParameterCountMismatch
	}

	argsRV := make([]reflect.Value, methodRT.NumIn())
	j := 0

	for i := range argsRV {
		if i == ccPos {
			argsRV[i] = reflect.ValueOf(cc)
			continue
		}
		if j >= len(args) {
			return nil, ErrMethodParameterCountMismatch
		}

		argRV, err := args[j].Convert(methodRT.In(i))
		if err != nil {
			return nil, ErrMethodParameterTypeMismatch
		}

		argsRV[i] = argRV
		j++
	}

	return argsRV, nil
}
