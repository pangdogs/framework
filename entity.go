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

// EntityBehavior 实体行为，在需要扩展实体能力时，匿名嵌入至实体结构体中
type EntityBehavior struct {
	ec.EntityBehavior
	logger      *zap.Logger
	sugarLogger *zap.SugaredLogger
}

// Runtime 获取运行时
func (e *EntityBehavior) Runtime() IRuntime {
	return reinterpret.Cast[IRuntime](runtime.Current(e))
}

// Service 获取服务
func (e *EntityBehavior) Service() IService {
	return reinterpret.Cast[IService](service.Current(e))
}

// L 结构化日志
func (e *EntityBehavior) L() *zap.Logger {
	if e.logger == nil {
		e.logger = log.L(e.Runtime()).With(zap.Any("entity", json.RawMessage(types.String2Bytes(e.String()))))
	}
	return e.logger
}

// S 传统日志
func (e *EntityBehavior) S() *zap.SugaredLogger {
	if e.sugarLogger == nil {
		e.sugarLogger = e.L().Sugar()
	}
	return e.sugarLogger
}
