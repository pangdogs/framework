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

package log

import (
	"git.golaxy.org/core/utils/option"
	"go.uber.org/zap"
)

// LoggerOptions 配置日志 add-in 使用的基础 Zap 日志器。
type LoggerOptions struct {
	Logger *zap.Logger // Logger 为 nil 时创建输出到标准输出的开发日志器。
}

// With 提供日志 add-in 的 Option 构造方法。
var With _Option

type _Option struct{}

// Default 返回使用内置开发日志器的默认设置。
func (_Option) Default() option.Setting[LoggerOptions] {
	return func(options *LoggerOptions) {
		With.Logger(nil)(options)
	}
}

// Logger 设置要派生 service 或 runtime 日志器的基础 Zap 日志器。
func (_Option) Logger(logger *zap.Logger) option.Setting[LoggerOptions] {
	return func(options *LoggerOptions) {
		options.Logger = logger
	}
}
