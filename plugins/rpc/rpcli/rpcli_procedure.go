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
	// OnewayRPC 单向RPC调用
	OnewayRPC(service, comp, method string, args ...any) error
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

// OnewayRPC 单向RPC调用
func (p *Procedure) OnewayRPC(service, comp, method string, args ...any) error {
	return p.cli.OnewayRPCToEntity(p.id, service, comp, method, args...)
}
