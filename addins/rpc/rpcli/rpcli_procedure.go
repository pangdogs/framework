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
	"errors"
	"fmt"
	"git.golaxy.org/core"
)

// AddProcedure 添加过程
func (c *RPCli) AddProcedure(name string, proc any) error {
	_proc, ok := proc.(IProcedure)
	if !ok {
		return fmt.Errorf("rpcli: %w: incorrect proc type", core.ErrArgs)
	}

	_proc.init(c, name, proc)
	cacheCallPath(name, _proc.GetReflected().Type())

	if !c.procs.TryAdd(name, _proc) {
		return ErrProcedureExists
	}

	return nil
}

// RemoveProcedure 删除过程
func (c *RPCli) RemoveProcedure(name string) error {
	if name == "" {
		return errors.New("rpcli: the main procedure can't be removed")
	}

	if !c.procs.Delete(name) {
		return ErrProcedureNotFound
	}

	return nil
}

// GetProcedure 查询过程
func (c *RPCli) GetProcedure(name string) (IProcedure, bool) {
	return c.procs.Get(name)
}
