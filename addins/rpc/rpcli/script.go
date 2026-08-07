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

// IScript 表示可由远端 RPC 调用的客户端脚本。
type IScript interface {
	iScript
	// Cli 返回脚本所属的 RPC 客户端。
	Cli() *RPCli
	// Name 返回脚本注册名称。
	Name() string
	// Reflected 返回脚本实例的反射值。
	Reflected() reflect.Value
}

type iScript interface {
	init(cli *RPCli, name string, instance any)
}

// ScriptBehavior 提供客户端脚本所需的基础实现。
type ScriptBehavior struct {
	cli       *RPCli
	name      string
	reflected reflect.Value
}

func (p *ScriptBehavior) init(cli *RPCli, name string, instance any) {
	p.cli = cli
	p.name = name
	p.reflected = reflect.ValueOf(instance)
}

// Cli 返回脚本所属的 RPC 客户端。
func (p *ScriptBehavior) Cli() *RPCli {
	return p.cli
}

// Name 返回脚本注册名称。
func (p *ScriptBehavior) Name() string {
	return p.name
}

// Reflected 返回脚本实例的反射值。
func (p *ScriptBehavior) Reflected() reflect.Value {
	return p.reflected
}
