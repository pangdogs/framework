package rpcli

import (
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/util/uid"
	"reflect"
)

var (
	Main = uid.Nil // 主过程
)

// IProc 过程接口
type IProc interface {
	_IProc

	GetCli() *RPCli
	GetId() uid.Id
	GetReflected() reflect.Value
	RPC(service, comp, method string, args ...any) runtime.AsyncRet
	OneWayRPC(service, comp, method string, args ...any) error
}

type _IProc interface {
	setup(cli *RPCli, entityId uid.Id, composite any)
}

// Proc 过程
type Proc struct {
	cli       *RPCli
	id        uid.Id
	reflected reflect.Value
}

func (p *Proc) setup(cli *RPCli, entityId uid.Id, composite any) {
	p.cli = cli
	p.id = entityId
	p.reflected = reflect.ValueOf(composite)
}

// GetCli 获取RPC客户端
func (p *Proc) GetCli() *RPCli {
	return p.cli
}

// GetId 获取实体Id
func (p *Proc) GetId() uid.Id {
	return p.id
}

// GetReflected 获取反射值
func (p *Proc) GetReflected() reflect.Value {
	return p.reflected
}

// RPC RPC调用
func (p *Proc) RPC(service, comp, method string, args ...any) runtime.AsyncRet {
	return p.cli.RPCToEntity(p.id, service, comp, method, args...)
}

// OneWayRPC 单向RPC调用
func (p *Proc) OneWayRPC(service, comp, method string, args ...any) error {
	return p.cli.OneWayRPCToEntity(p.id, service, comp, method, args...)
}
