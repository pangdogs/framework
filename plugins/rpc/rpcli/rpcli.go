package rpcli

import (
	"errors"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/codec"
	"git.golaxy.org/framework/net/gap/variant"
	"git.golaxy.org/framework/plugins/gate/cli"
	"git.golaxy.org/framework/plugins/rpc/callpath"
	"git.golaxy.org/framework/utils/concurrent"
)

var (
	ErrEntityNotFound               = errors.New("rpc: entity not found")                // 找不到实体
	ErrMethodNotFound               = errors.New("rpc: method not found")                // 找不到方法
	ErrMethodParameterCountMismatch = errors.New("rpc: method parameter count mismatch") // 方法参数数量不匹配
	ErrMethodParameterTypeMismatch  = errors.New("rpc: method parameter type mismatch")  // 方法参数类型不匹配
)

// RPCli RCP客户端
type RPCli struct {
	*cli.Client
	encoder codec.Encoder
	decoder codec.Decoder
	procs   concurrent.LockedMap[uid.Id, IProcedure]
}

// RPC RPC调用
func (c *RPCli) RPC(service, comp, method string, args ...any) async.AsyncRet {
	return c.RPCToEntity(uid.Nil, service, comp, method, args...)
}

// RPCToEntity 实体RPC调用
func (c *RPCli) RPCToEntity(entityId uid.Id, service, comp, method string, args ...any) async.AsyncRet {
	ret := concurrent.MakeRespAsyncRet()
	future := concurrent.MakeFuture(c.GetFutures(), nil, ret)

	vargs, err := variant.MakeArrayReadonly(args)
	if err != nil {
		future.Cancel(err)
		return ret.ToAsyncRet()
	}

	cp := callpath.CallPath{
		Category:  callpath.Entity,
		EntityId:  entityId,
		Component: comp,
		Method:    method,
	}

	msg := &gap.MsgRPCRequest{
		CorrId: future.Id,
		Path:   cp.String(),
		Args:   vargs,
	}

	msgbs, err := gap.Marshal(msg)
	if err != nil {
		future.Cancel(err)
		return ret.ToAsyncRet()
	}
	defer msgbs.Release()

	forwardMsg := &gap.MsgForward{
		Dst:       service,
		CorrId:    msg.CorrId,
		TransId:   msg.MsgId(),
		TransData: msgbs.Data(),
	}

	bs, err := c.encoder.EncodeBytes("", 0, forwardMsg)
	if err != nil {
		future.Cancel(err)
		return ret.ToAsyncRet()
	}
	defer bs.Release()

	if err = c.SendData(bs.Data()); err != nil {
		future.Cancel(err)
		return ret.ToAsyncRet()
	}

	return ret.ToAsyncRet()
}

// OneWayRPC 单向RPC调用
func (c *RPCli) OneWayRPC(service, comp, method string, args ...any) error {
	return c.OneWayRPCToEntity(uid.Nil, service, comp, method, args...)
}

// OneWayRPCToEntity 实体单向RPC调用
func (c *RPCli) OneWayRPCToEntity(entityId uid.Id, service, comp, method string, args ...any) error {
	vargs, err := variant.MakeArrayReadonly(args)
	if err != nil {
		return err
	}

	cp := callpath.CallPath{
		Category:  callpath.Entity,
		EntityId:  entityId,
		Component: comp,
		Method:    method,
	}

	msg := &gap.MsgOneWayRPC{
		Path: cp.String(),
		Args: vargs,
	}

	msgbs, err := gap.Marshal(msg)
	if err != nil {
		return err
	}
	defer msgbs.Release()

	forwardMsg := &gap.MsgForward{
		Dst:       service,
		TransId:   msg.MsgId(),
		TransData: msgbs.Data(),
	}

	bs, err := c.encoder.EncodeBytes("", 0, forwardMsg)
	if err != nil {
		return err
	}
	defer bs.Release()

	if err = c.SendData(bs.Data()); err != nil {
		return err
	}

	return nil
}

// AddProcedure 添加过程
func (c *RPCli) AddProcedure(id uid.Id, proc any) error {
	if id.IsNil() {
		return fmt.Errorf("%w: id is nil", core.ErrArgs)
	}

	_proc, ok := proc.(IProcedure)
	if !ok {
		return fmt.Errorf("%w: incorrect proc type", core.ErrArgs)
	}

	_proc.init(c, id, proc)
	c.procs.Add(id, _proc)

	return nil
}

// RemoveProcedure 删除过程
func (c *RPCli) RemoveProcedure(id uid.Id) error {
	if id.IsNil() {
		return fmt.Errorf("%w: id is nil", core.ErrArgs)
	}

	c.procs.Delete(id)

	return nil
}

// GetProcedure 查询过程
func (c *RPCli) GetProcedure(id uid.Id) (IProcedure, bool) {
	return c.procs.Get(id)
}
