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
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/framework/addins/rpc/callpath"
	"git.golaxy.org/framework/addins/rpcstack"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/variant"
	"reflect"
	"time"
)

var (
	callChainRT = reflect.TypeFor[rpcstack.CallChain]()
)

func (c *RPCli) handleRecvData(data []byte) error {
	mp, err := c.decoder.Decode(data)
	if err != nil {
		return err
	}

	switch mp.Head.MsgId {
	case gap.MsgId_OnewayRPC:
		return c.acceptNotify(mp.Head.Src, mp.Msg.(*gap.MsgOnewayRPC))

	case gap.MsgId_RPC_Request:
		return c.acceptRequest(mp.Head.Src, mp.Msg.(*gap.MsgRPCRequest))

	case gap.MsgId_RPC_Reply:
		return c.resolve(mp.Msg.(*gap.MsgRPCReply))
	}

	return nil
}

func (c *RPCli) acceptNotify(src gap.Origin, req *gap.MsgOnewayRPC) error {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		return fmt.Errorf("rpcli: parse rpc notify path:%q failed, %s", req.Path, err)
	}

	cc := append(req.CallChain, rpcstack.Call{Svc: src.Svc, Addr: src.Addr, Timestamp: time.UnixMilli(src.Timestamp).Local(), Transit: true})

	switch cp.Category {
	case callpath.Client:
		if rets, err := c.callProc(cc, cp.Script, cp.Method, req.Args); err != nil {
			c.GetLogger().Errorf("rpc notify entity:%q, method:%q calls failed, %s", cp.Id, cp.Method, err)
		} else {
			c.GetLogger().Debugf("rpc notify entity:%q, method:%q calls finished", cp.Id, cp.Method)
			rets.Release()
		}
		return nil
	}

	return nil
}

func (c *RPCli) acceptRequest(src gap.Origin, req *gap.MsgRPCRequest) error {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		err = fmt.Errorf("rpcli: parse rpc request(%d) path %q failed, %s", req.CorrId, req.Path, err)
		go c.reply(src, req.CorrId, nil, err)
		return err
	}

	cc := append(req.CallChain, rpcstack.Call{Svc: src.Svc, Addr: src.Addr, Timestamp: time.UnixMilli(src.Timestamp).Local(), Transit: true})

	switch cp.Category {
	case callpath.Client:
		rets, err := c.callProc(cc, cp.Script, cp.Method, req.Args)
		if err != nil {
			c.GetLogger().Errorf("rpc request(%d) entity:%q, method:%q calls failed, %s", req.CorrId, cp.Id, cp.Method, err)
		} else {
			c.GetLogger().Debugf("rpc request(%d) entity:%q, method:%q calls finished", req.CorrId, cp.Id, cp.Method)
		}
		go c.reply(src, req.CorrId, rets, err)
		return nil
	}

	return nil
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
		msg.Error = *variant.MakeError(retErr)
	}

	msgBuf, err := gap.Marshal(msg)
	if err != nil {
		c.GetLogger().Errorf("rpc reply(%d) to src:%q failed, %s", corrId, src.Addr, err)
		return
	}
	defer msgBuf.Release()

	forwardMsg := &gap.MsgForward{
		Dst:       src.Addr,
		CorrId:    msg.CorrId,
		TransId:   msg.MsgId(),
		TransData: msgBuf.Data(),
	}

	mpBuf, err := c.encoder.Encode(gap.Origin{Timestamp: c.remoteTime.NowTime().UnixMilli()}, 0, forwardMsg)
	if err != nil {
		c.GetLogger().Errorf("rpc reply(%d) to src:%q failed, %s", corrId, src.Addr, err)
		return
	}
	defer mpBuf.Release()

	if err = c.SendData(mpBuf.Data()); err != nil {
		c.GetLogger().Errorf("rpc reply(%d) to src:%q failed, %s", corrId, src.Addr, err)
		return
	}

	c.GetLogger().Debugf("rpc reply(%d) to src:%q ok", corrId, src.Addr)
}

func (c *RPCli) resolve(reply *gap.MsgRPCReply) error {
	ret := async.Ret{}

	if reply.Error.OK() {
		if len(reply.Rets) > 0 {
			ret.Value = reply.Rets
		}
	} else {
		ret.Error = &reply.Error
	}

	return c.GetFutures().Resolve(reply.CorrId, ret)
}

func (c *RPCli) callProc(cc rpcstack.CallChain, procedure, method string, args variant.Array) (rets variant.Array, err error) {
	proc, ok := c.procs.Get(procedure)
	if !ok {
		return nil, ErrProcedureNotFound
	}

	methodRV := proc.GetReflected().MethodByName(method)
	if !methodRV.IsValid() {
		return nil, ErrMethodNotFound
	}

	argsRV, err := parseArgs(methodRV, cc, args)
	if err != nil {
		return nil, err
	}

	return variant.MakeSerializedArray(methodRV.Call(argsRV))
}

func parseArgs(methodRV reflect.Value, cc rpcstack.CallChain, args variant.Array) ([]reflect.Value, error) {
	methodRT := methodRV.Type()
	var argsRV []reflect.Value
	var argsPos int

	switch methodRT.NumIn() {
	case len(args) + 1:
		if !callChainRT.AssignableTo(methodRT.In(0)) {
			return nil, ErrMethodParameterTypeMismatch
		}
		argsRV = append(make([]reflect.Value, 0, len(args)+1), reflect.ValueOf(cc))
		argsPos = 1

	case len(args):
		argsRV = make([]reflect.Value, 0, len(args))
		argsPos = 0

	default:
		return nil, ErrMethodParameterCountMismatch
	}

	for i := range args {
		argRV, err := args[i].Convert(methodRT.In(argsPos + i))
		if err != nil {
			return nil, ErrMethodParameterTypeMismatch
		}
		argsRV = append(argsRV, argRV)
	}

	return argsRV, nil
}
