package rpcli

import (
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/uid"
	"reflect"
)

var (
	Main = uid.Nil // 主过程
)

// IProcedure 过程接口
type IProcedure interface {
	iProcedure
	// GetCli 获取RPC客户端
	GetCli() *RPCli
	// GetId 获取实体Id
	GetId() uid.Id
	// GetReflected 获取反射值
	GetReflected() reflect.Value
	// RPC RPC调用
	RPC(service, comp, method string, args ...any) async.AsyncRet
	// OneWayRPC 单向RPC调用
	OneWayRPC(service, comp, method string, args ...any) error
}

type iProcedure interface {
	init(cli *RPCli, entityId uid.Id, composite any)
}

// Procedure 过程
type Procedure struct {
	cli       *RPCli
	id        uid.Id
	reflected reflect.Value
}

func (p *Procedure) init(cli *RPCli, entityId uid.Id, composite any) {
	p.cli = cli
	p.id = entityId
	p.reflected = reflect.ValueOf(composite)
}

// GetCli 获取RPC客户端
func (p *Procedure) GetCli() *RPCli {
	return p.cli
}

// GetId 获取实体Id
func (p *Procedure) GetId() uid.Id {
	return p.id
}

// GetReflected 获取反射值
func (p *Procedure) GetReflected() reflect.Value {
	return p.reflected
}

// RPC RPC调用
func (p *Procedure) RPC(service, comp, method string, args ...any) async.AsyncRet {
	return p.cli.RPCToEntity(p.id, service, comp, method, args...)
}

// OneWayRPC 单向RPC调用
func (p *Procedure) OneWayRPC(service, comp, method string, args ...any) error {
	return p.cli.OneWayRPCToEntity(p.id, service, comp, method, args...)
}
