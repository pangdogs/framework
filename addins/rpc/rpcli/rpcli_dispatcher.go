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
	"git.golaxy.org/framework/utils/correlation"
	"go.uber.org/zap"
)

var (
	callChainRT = reflect.TypeFor[rpcstack.CallChain]()
)

// handleData 解码客户端收到的 GAP 数据，并按消息类型分发 RPC。
func (c *RPCli) handleData(data []byte) {
	mp, err := c.decoder.Decode(data)
	if err != nil {
		c.L().Error("decode data failed",
			zap.String("session_id", c.SessionID().String()),
			zap.Error(err))
		return
	}

	switch mp.Head.MsgID {
	case gap.MsgID_OnewayRPC:
		c.acceptNotify(mp.Head.Src, mp.Body.(*gap.MsgOnewayRPC))

	case gap.MsgID_RPC_Request:
		c.acceptRequest(mp.Head.Src, mp.Body.(*gap.MsgRPCRequest))

	case gap.MsgID_RPC_Reply:
		c.resolveReply(mp.Body.(*gap.MsgRPCReply))
	}
}

// acceptNotify 校验调用路径并同步调用本地客户端脚本，不发送响应。
func (c *RPCli) acceptNotify(src gap.Origin, req *gap.MsgOnewayRPC) {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		c.L().Error("accept rpc notify failed",
			zap.String("session_id", c.SessionID().String()),
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
		_, err := c.callScript(cc, cp.Script, cp.Method, req.Args)
		if err != nil {
			c.L().Error("accept rpc notify failed",
				zap.String("session_id", c.SessionID().String()),
				zap.String("local", c.NetAddr().Local.String()),
				zap.String("remote", c.NetAddr().Remote.String()),
				zap.String("script", cp.Script),
				zap.String("method", cp.Method),
				zap.Error(err))
		} else {
			c.L().Debug("accept rpc notify finished",
				zap.String("session_id", c.SessionID().String()),
				zap.String("local", c.NetAddr().Local.String()),
				zap.String("remote", c.NetAddr().Remote.String()),
				zap.String("call_path", cp.String()),
				zap.String("script", cp.Script),
				zap.String("method", cp.Method))
		}
	}
}

// acceptRequest 调用本地客户端脚本，并将返回值或错误回复给请求来源。
func (c *RPCli) acceptRequest(src gap.Origin, req *gap.MsgRPCRequest) {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		err = fmt.Errorf("parse call path failed: %w", err)
		c.L().Error("accept rpc request failed",
			zap.String("session_id", c.SessionID().String()),
			zap.String("local", c.NetAddr().Local.String()),
			zap.String("remote", c.NetAddr().Remote.String()),
			zap.Uint64("corr_id", uint64(req.CorrID)),
			zap.Error(err))
		c.reply(src, req.CorrID, variant.Array{}, err)
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
			c.L().Error("accept rpc request failed",
				zap.String("session_id", c.SessionID().String()),
				zap.String("local", c.NetAddr().Local.String()),
				zap.String("remote", c.NetAddr().Remote.String()),
				zap.Uint64("corr_id", uint64(req.CorrID)),
				zap.String("call_path", cp.String()),
				zap.String("script", cp.Script),
				zap.String("method", cp.Method),
				zap.Error(err))
		} else {
			c.L().Debug("accept rpc request finished",
				zap.String("session_id", c.SessionID().String()),
				zap.String("local", c.NetAddr().Local.String()),
				zap.String("remote", c.NetAddr().Remote.String()),
				zap.Uint64("corr_id", uint64(req.CorrID)),
				zap.String("call_path", cp.String()),
				zap.String("script", cp.Script),
				zap.String("method", cp.Method))
		}
		c.reply(src, req.CorrID, rets, err)
	}
}

// resolveReply 按关联 ID 完成客户端发起请求时创建的 Future。
func (c *RPCli) resolveReply(reply *gap.MsgRPCReply) {
	ret := async.Result{}

	if reply.Error.OK() {
		if len(reply.Rets.Items) > 0 {
			ret.Value = reply.Rets
		}
	} else {
		ret.Error = &reply.Error
	}

	if !c.Correlation().Resolve(reply.CorrID, ret) {
		c.L().Error("resolve rpc reply failed",
			zap.String("session_id", c.SessionID().String()),
			zap.String("local", c.NetAddr().Local.String()),
			zap.String("remote", c.NetAddr().Remote.String()),
			zap.Uint64("corr_id", uint64(reply.CorrID)))
		return
	}

	c.L().Debug("rpc reply resolved",
		zap.String("session_id", c.SessionID().String()),
		zap.String("local", c.NetAddr().Local.String()),
		zap.String("remote", c.NetAddr().Remote.String()),
		zap.Uint64("corr_id", uint64(reply.CorrID)))
}

// reply 将脚本调用结果包装为转发消息并发回来源地址；零关联 ID 不回复。
func (c *RPCli) reply(src gap.Origin, corrID correlation.ID, rets variant.Array, retErr error) {
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

	msgBuf, err := gap.Marshal(msg)
	if err != nil {
		c.L().Error("marshal rpc reply failed",
			zap.String("session_id", c.SessionID().String()),
			zap.String("local", c.NetAddr().Local.String()),
			zap.String("remote", c.NetAddr().Remote.String()),
			zap.Uint64("corr_id", uint64(corrID)),
			zap.Error(err))
		return
	}
	defer msgBuf.Release()

	forwardMsg := &gap.MsgForward{
		Dst:       src.Addr,
		CorrID:    msg.CorrID,
		TransID:   msg.MsgID(),
		TransData: msgBuf.Payload(),
	}

	mpBuf, err := c.encoder.Encode(gap.Origin{Timestamp: c.remoteClock.RemoteNow().UnixMilli()}, 0, forwardMsg)
	if err != nil {
		c.L().Error("encode rpc reply failed",
			zap.String("session_id", c.SessionID().String()),
			zap.String("local", c.NetAddr().Local.String()),
			zap.String("remote", c.NetAddr().Remote.String()),
			zap.Uint64("corr_id", uint64(corrID)),
			zap.Error(err))
		return
	}
	defer mpBuf.Release()

	if err = c.DataIO().Send(mpBuf.Payload()); err != nil {
		c.L().Error("send rpc reply failed",
			zap.String("session_id", c.SessionID().String()),
			zap.String("local", c.NetAddr().Local.String()),
			zap.String("remote", c.NetAddr().Remote.String()),
			zap.Uint64("corr_id", uint64(corrID)),
			zap.Error(err))
		return
	}

	c.L().Debug("rpc reply sent",
		zap.String("session_id", c.SessionID().String()),
		zap.String("local", c.NetAddr().Local.String()),
		zap.String("remote", c.NetAddr().Remote.String()),
		zap.Uint64("corr_id", uint64(corrID)))
}

// callScript 查找已注册脚本和导出方法，转换参数后通过反射同步调用。
func (c *RPCli) callScript(cc rpcstack.CallChain, script, method string, args variant.Array) (rets variant.Array, err error) {
	scr, ok := c.GetScript(script)
	if !ok {
		return variant.Array{}, ErrScriptNotFound
	}

	methodRV := scr.Reflected().MethodByName(method)
	if !methodRV.IsValid() {
		return variant.Array{}, ErrMethodNotFound
	}

	argsRV, err := parseArgs(methodRV, cc, args)
	if err != nil {
		return variant.Array{}, err
	}

	return variant.NewArray(methodRV.Call(argsRV))
}

// parseArgs 将协议参数转换为方法参数，并在唯一显式声明的 CallChain 位置注入调用链。
func parseArgs(methodRV reflect.Value, cc rpcstack.CallChain, args variant.Array) ([]reflect.Value, error) {
	methodRT := methodRV.Type()
	ccPos := -1

	for i := range methodRT.NumIn() {
		if methodRT.In(i) != callChainRT {
			continue
		}
		if ccPos >= 0 {
			return nil, ErrMethodParameterCountMismatch
		}
		ccPos = i
	}

	switch {
	case ccPos < 0 && methodRT.NumIn() != len(args.Items):
		return nil, ErrMethodParameterCountMismatch
	case ccPos >= 0 && methodRT.NumIn() != len(args.Items)+1:
		return nil, ErrMethodParameterCountMismatch
	}

	argsRV := make([]reflect.Value, methodRT.NumIn())
	j := 0

	for i := range argsRV {
		if i == ccPos {
			argsRV[i] = reflect.ValueOf(cc)
			continue
		}
		if j >= len(args.Items) {
			return nil, ErrMethodParameterCountMismatch
		}

		argRV, err := args.Items[j].ToNative(methodRT.In(i))
		if err != nil {
			return nil, ErrMethodParameterTypeMismatch
		}

		argsRV[i] = argRV
		j++
	}

	return argsRV, nil
}
