package rpcli

import (
	"errors"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/util/uid"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/codec"
	"git.golaxy.org/framework/net/gap/variant"
	"git.golaxy.org/framework/plugins/gate/cli"
	"git.golaxy.org/framework/plugins/rpc/callpath"
	"git.golaxy.org/framework/util/concurrent"
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
	procs   concurrent.LockedMap[uid.Id, IProc]
}

// RPC RPC调用
func (c *RPCli) RPC(entityId uid.Id, service, comp, method string, args ...any) runtime.AsyncRet {
	return c.RPCToEntity(uid.Nil, service, comp, method, args...)
}

// RPCToEntity 实体RPC调用
func (c *RPCli) RPCToEntity(entityId uid.Id, service, comp, method string, args ...any) runtime.AsyncRet {
	ret := concurrent.MakeRespAsyncRet()
	future := concurrent.MakeFuture(c.GetFutures(), nil, ret)

	vargs, err := variant.MakeArray(args)
	if err != nil {
		future.Cancel(err)
		return ret.CastAsyncRet()
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
		return ret.CastAsyncRet()
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
		return ret.CastAsyncRet()
	}
	defer bs.Release()

	if err = c.SendData(bs.Data()); err != nil {
		future.Cancel(err)
		return ret.CastAsyncRet()
	}

	return ret.CastAsyncRet()
}

// OneWayRPC 单向RPC调用
func (c *RPCli) OneWayRPC(service, comp, method string, args ...any) error {
	return c.OneWayRPCToEntity(uid.Nil, service, comp, method, args...)
}

// OneWayRPCToEntity 实体单向RPC调用
func (c *RPCli) OneWayRPCToEntity(entityId uid.Id, service, comp, method string, args ...any) error {
	vargs, err := variant.MakeArray(args)
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

// AddProc 添加过程
func (c *RPCli) AddProc(id uid.Id, proc any) error {
	if id == Main {
		return fmt.Errorf("%w: id is nil", core.ErrArgs)
	}

	_proc, ok := proc.(IProc)
	if !ok {
		return fmt.Errorf("%w: incorrect proc type", core.ErrArgs)
	}

	_proc.setup(c, id, proc)
	c.procs.Insert(id, _proc)

	return nil
}

// RemoveProc 删除过程
func (c *RPCli) RemoveProc(id uid.Id) error {
	if id.IsNil() {
		return fmt.Errorf("%w: id is nil", core.ErrArgs)
	}

	c.procs.Delete(id)

	return nil
}

// GetProc 查询过程
func (c *RPCli) GetProc(id uid.Id) (IProc, bool) {
	return c.procs.Get(id)
}
