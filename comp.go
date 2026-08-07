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

package framework

import (
	"encoding/json"

	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/reinterpret"
	"git.golaxy.org/core/utils/types"
	"git.golaxy.org/framework/addins/log"
	"go.uber.org/zap"
)

// ComponentBehavior 扩展 core 组件行为，供自定义组件匿名嵌入。
// 日志器会按组件实例惰性创建，因此应在组件所属运行时 goroutine 中使用。
type ComponentBehavior struct {
	ec.ComponentBehavior
	logger      *zap.Logger
	sugarLogger *zap.SugaredLogger
}

// Runtime 返回组件所属的 framework 运行时。
func (c *ComponentBehavior) Runtime() IRuntime {
	return reinterpret.Cast[IRuntime](runtime.Current(c))
}

// Service 返回承载组件所属运行时的服务。
func (c *ComponentBehavior) Service() IService {
	return reinterpret.Cast[IService](service.Current(c))
}

// L 返回附带当前组件字段的结构化日志器。
func (c *ComponentBehavior) L() *zap.Logger {
	if c.logger == nil {
		c.logger = log.L(c.Runtime()).With(zap.Any("component", json.RawMessage(types.String2Bytes(c.String()))))
	}
	return c.logger
}

// S 返回附带当前组件字段的 SugaredLogger。
func (c *ComponentBehavior) S() *zap.SugaredLogger {
	if c.sugarLogger == nil {
		c.sugarLogger = c.L().Sugar()
	}
	return c.sugarLogger
}
