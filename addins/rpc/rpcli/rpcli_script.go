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

	"git.golaxy.org/core/utils/types"
	"go.uber.org/zap"
)

// SetScripts 设置脚本
func (c *RPCli) SetScripts(scripts map[string]IScript) {
	c.scriptsMu.Lock()
	defer c.scriptsMu.Unlock()

	for name, script := range scripts {
		if script != nil {
			script.init(c, name, script)
			cacheCallPath(name, script.Reflected().Type())

			scriptType := script.Reflected().Type()
			for scriptType.Kind() != reflect.Pointer {
				scriptType = scriptType.Elem()
			}

			c.scripts.Add(name, script)
			c.Logger().Info("script added",
				zap.String("session_id", c.SessionId().String()),
				zap.String("name", name),
				zap.String("instance", types.FullNameRT(scriptType)))

		} else {
			if c.scripts.Delete(name) {
				c.Logger().Info("script removed",
					zap.String("session_id", c.SessionId().String()),
					zap.String("name", name))
			}
		}
	}
}

// GetScript 查询脚本
func (c *RPCli) GetScript(name string) (IScript, bool) {
	c.scriptsMu.RLock()
	defer c.scriptsMu.RUnlock()

	return c.scripts.Get(name)
}
