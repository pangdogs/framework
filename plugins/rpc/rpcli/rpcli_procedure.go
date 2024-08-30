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
	"reflect"
)

// IProcedure 过程接口
type IProcedure interface {
	iProcedure
	// GetCli 获取RPC客户端
	GetCli() *RPCli
	// GetName 获取名称
	GetName() string
	// GetReflected 获取反射值
	GetReflected() reflect.Value
}

type iProcedure interface {
	init(cli *RPCli, name string, instance any)
}

// Procedure 过程
type Procedure struct {
	cli       *RPCli
	name      string
	reflected reflect.Value
}

func (p *Procedure) init(cli *RPCli, name string, instance any) {
	p.cli = cli
	p.name = name
	p.reflected = reflect.ValueOf(instance)
}

// GetCli 获取RPC客户端
func (p *Procedure) GetCli() *RPCli {
	return p.cli
}

// GetName 获取名称
func (p *Procedure) GetName() string {
	return p.name
}

// GetReflected 获取反射值
func (p *Procedure) GetReflected() reflect.Value {
	return p.reflected
}
