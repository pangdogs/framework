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
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/variant"
	"git.golaxy.org/framework/plugins/rpc/callpath"
	"reflect"
)

func (c *RPCli) handleRecvData(data []byte) error {
	mp, err := c.decoder.Decode(data)
	if err != nil {
		return err
	}

	switch mp.Head.MsgId {
	case gap.MsgId_OneWayRPC:
		return c.acceptNotify(mp.Msg.(*gap.MsgOneWayRPC))

	case gap.MsgId_RPC_Request:
		return c.acceptRequest(mp.Head.Src, mp.Msg.(*gap.MsgRPCRequest))

	case gap.MsgId_RPC_Reply:
		return c.resolve(mp.Msg.(*gap.MsgRPCReply))
	}

	return nil
}

func (c *RPCli) acceptNotify(req *gap.MsgOneWayRPC) error {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		return fmt.Errorf("parse rpc notify path:%q failed, %s", req.Path, err)
	}

	switch cp.Category {
	case callpath.Client:
		if rets, err := c.callProc(cp.EntityId, cp.Method, req.Args); err != nil {
			c.GetLogger().Errorf("rpc notify entity:%q, method:%q calls failed, %s", cp.EntityId, cp.Method, err)
		} else {
			c.GetLogger().Debugf("rpc notify entity:%q, method:%q calls finished", cp.EntityId, cp.Method)
			rets.Release()
		}
		return nil
	}

	return nil
}

func (c *RPCli) acceptRequest(src string, req *gap.MsgRPCRequest) error {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		err = fmt.Errorf("parse rpc request(%d) path %q failed, %s", req.CorrId, req.Path, err)
		go c.reply(src, req.CorrId, nil, err)
		return err
	}

	switch cp.Category {
	case callpath.Client:
		rets, err := c.callProc(cp.EntityId, cp.Method, req.Args)
		if err != nil {
			c.GetLogger().Errorf("rpc request(%d) entity:%q, method:%q calls failed, %s", req.CorrId, cp.EntityId, cp.Method, err)
		} else {
			c.GetLogger().Debugf("rpc request(%d) entity:%q, method:%q calls finished", req.CorrId, cp.EntityId, cp.Method)
		}
		go c.reply(src, req.CorrId, rets, err)
		return nil
	}

	return nil
}

func (c *RPCli) reply(src string, corrId int64, rets variant.Array, retErr error) {
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
		c.GetLogger().Errorf("rpc reply(%d) to src:%q failed, %s", corrId, src, err)
		return
	}
	defer msgBuf.Release()

	forwardMsg := &gap.MsgForward{
		Dst:       src,
		CorrId:    msg.CorrId,
		TransId:   msg.MsgId(),
		TransData: msgBuf.Data(),
	}

	mpBuf, err := c.encoder.Encode("", "", 0, forwardMsg)
	if err != nil {
		c.GetLogger().Errorf("rpc reply(%d) to src:%q failed, %s", corrId, src, err)
		return
	}
	defer mpBuf.Release()

	if err = c.SendData(mpBuf.Data()); err != nil {
		c.GetLogger().Errorf("rpc reply(%d) to src:%q failed, %s", corrId, src, err)
		return
	}

	c.GetLogger().Debugf("rpc reply(%d) to src:%q ok", corrId, src)
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

func (c *RPCli) callProc(entityId uid.Id, method string, args variant.Array) (rets variant.Array, err error) {
	proc, ok := c.procs.Get(entityId)
	if !ok {
		return nil, ErrEntityNotFound
	}

	methodRV := proc.GetReflected().MethodByName(method)
	if !methodRV.IsValid() {
		return nil, ErrMethodNotFound
	}

	argsRV, err := parseArgs(methodRV, args)
	if err != nil {
		return nil, err
	}

	return variant.MakeSerializedArray(methodRV.Call(argsRV))
}

func parseArgs(methodRV reflect.Value, args variant.Array) ([]reflect.Value, error) {
	methodRT := methodRV.Type()

	if methodRT.NumIn() != len(args) {
		return nil, ErrMethodParameterCountMismatch
	}

	argsRV := make([]reflect.Value, 0, len(args))

	for i := range args {
		argRV, err := args[i].Convert(methodRT.In(i))
		if err != nil {
			return nil, ErrMethodParameterTypeMismatch
		}
		argsRV = append(argsRV, argRV)
	}

	return argsRV, nil
}
