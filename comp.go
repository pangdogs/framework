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
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/reinterpret"
	"git.golaxy.org/framework/addins/log"
	"go.uber.org/zap"
)

// ComponentBehavior 组件行为，在开发新组件时，匿名嵌入至组件结构体中
type ComponentBehavior struct {
	ec.ComponentBehavior
	logger      *zap.Logger
	sugarLogger *zap.SugaredLogger
}

// Runtime 获取运行时
func (c *ComponentBehavior) Runtime() IRuntime {
	return reinterpret.Cast[IRuntime](runtime.Current(c))
}

// Service 获取服务
func (c *ComponentBehavior) Service() IService {
	return reinterpret.Cast[IService](service.Current(c))
}

// L 结构化日志
func (c *ComponentBehavior) L() *zap.Logger {
	if c.logger == nil {
		c.logger = log.L(c.Runtime()).With(zap.String("component", c.String()))
	}
	return c.logger
}

// S 传统日志
func (c *ComponentBehavior) S() *zap.SugaredLogger {
	if c.sugarLogger == nil {
		c.sugarLogger = c.L().Sugar()
	}
	return c.sugarLogger
}
