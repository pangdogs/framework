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

// EntityBehavior 扩展 core 实体行为，供自定义实体匿名嵌入。
// 日志器会按实体实例惰性创建，因此应在实体所属运行时 goroutine 中使用。
type EntityBehavior struct {
	ec.EntityBehavior
	logger      *zap.Logger
	sugarLogger *zap.SugaredLogger
}

// Runtime 返回实体所属的 framework 运行时。
func (e *EntityBehavior) Runtime() IRuntime {
	return reinterpret.Cast[IRuntime](runtime.Current(e))
}

// Service 返回承载实体所属运行时的服务。
func (e *EntityBehavior) Service() IService {
	return reinterpret.Cast[IService](service.Current(e))
}

// L 返回附带当前实体字段的结构化日志器。
func (e *EntityBehavior) L() *zap.Logger {
	if e.logger == nil {
		e.logger = log.L(e.Runtime()).With(zap.Any("entity", json.RawMessage(types.String2Bytes(e.String()))))
	}
	return e.logger
}

// S 返回附带当前实体字段的 SugaredLogger。
func (e *EntityBehavior) S() *zap.SugaredLogger {
	if e.sugarLogger == nil {
		e.sugarLogger = e.L().Sugar()
	}
	return e.sugarLogger
}
