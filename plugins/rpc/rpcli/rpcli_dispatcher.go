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
	mp, err := c.decoder.DecodeBytes(data)
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
		if _, err := c.callProc(cp.EntityId, cp.Method, req.Args); err != nil {
			c.GetLogger().Errorf("rpc notify entity:%q, method:%q calls failed, %s", cp.EntityId, cp.Method, err)
		} else {
			c.GetLogger().Debugf("rpc notify entity:%q, method:%q calls finished", cp.EntityId, cp.Method)
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
		retsRV, err := c.callProc(cp.EntityId, cp.Method, req.Args)
		if err != nil {
			c.GetLogger().Errorf("rpc request(%d) entity:%q, method:%q calls failed, %s", req.CorrId, cp.EntityId, cp.Method, err)
		} else {
			c.GetLogger().Debugf("rpc request(%d) entity:%q, method:%q calls finished", req.CorrId, cp.EntityId, cp.Method)
		}
		go c.reply(src, req.CorrId, retsRV, err)
		return nil
	}

	return nil
}

func (c *RPCli) reply(src string, corrId int64, retsRV []reflect.Value, retErr error) {
	if corrId == 0 {
		return
	}

	msg := &gap.MsgRPCReply{
		CorrId: corrId,
	}

	var err error
	msg.Rets, err = variant.MakeArrayReadonly(retsRV)
	if err != nil {
		c.GetLogger().Errorf("rpc reply(%d) to src:%q failed, %s", corrId, src, err)
		return
	}

	if retErr != nil {
		msg.Error = *variant.MakeError(retErr)
	}

	msgbs, err := gap.Marshal(msg)
	if err != nil {
		c.GetLogger().Errorf("rpc reply(%d) to src:%q failed, %s", corrId, src, err)
		return
	}
	defer msgbs.Release()

	forwardMsg := &gap.MsgForward{
		Dst:       src,
		CorrId:    msg.CorrId,
		TransId:   msg.MsgId(),
		TransData: msgbs.Data(),
	}

	bs, err := c.encoder.EncodeBytes("", 0, forwardMsg)
	if err != nil {
		c.GetLogger().Errorf("rpc reply(%d) to src:%q failed, %s", corrId, src, err)
		return
	}
	defer bs.Release()

	if err = c.SendData(bs.Data()); err != nil {
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

func (c *RPCli) callProc(entityId uid.Id, method string, args variant.Array) (rets []reflect.Value, err error) {
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

	return methodRV.Call(argsRV), nil
}

func parseArgs(methodRV reflect.Value, args variant.Array) ([]reflect.Value, error) {
	methodRT := methodRV.Type()

	if methodRT.NumIn() != len(args) {
		return nil, ErrMethodParameterCountMismatch
	}

	argsRV := make([]reflect.Value, 0, len(args))

	for i := range args {
		argRV := args[i].Reflected
		argRT := argRV.Type()
		inRT := methodRT.In(i)

	retry:
		if argRT.AssignableTo(inRT) {
			argsRV = append(argsRV, argRV)
			continue
		}

		if argRV.CanConvert(inRT) {
			if argRT.Size() > inRT.Size() {
				return nil, ErrMethodParameterTypeMismatch
			}
			argsRV = append(argsRV, argRV.Convert(inRT))
			continue
		}

		if argRT.Kind() == reflect.Pointer {
			argRV = argRV.Elem()
			argRT = argRV.Type()
			goto retry
		}

		argRV, err := variant.CastVariantReflected(args[i], inRT)
		if err != nil {
			return nil, ErrMethodParameterTypeMismatch
		}

		argsRV = append(argsRV, argRV)
	}

	return argsRV, nil
}
