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

// LoggerOptions 所有选项
type LoggerOptions struct {
	Logger      *zap.Logger
	ServiceInfo bool
	RuntimeInfo bool
}

var With _Option

type _Option struct{}

// Default 默认值
func (_Option) Default() option.Setting[LoggerOptions] {
	return func(options *LoggerOptions) {
		With.Logger(nil)(options)
		With.ServiceInfo(true)(options)
		With.RuntimeInfo(true)(options)
	}
}

// Logger 日志
func (_Option) Logger(logger *zap.Logger) option.Setting[LoggerOptions] {
	return func(options *LoggerOptions) {
		options.Logger = logger
	}
}

// ServiceInfo 添加服务信息
func (_Option) ServiceInfo(b bool) option.Setting[LoggerOptions] {
	return func(options *LoggerOptions) {
		options.ServiceInfo = b
	}
}

// RuntimeInfo 添加运行时信息
func (_Option) RuntimeInfo(b bool) option.Setting[LoggerOptions] {
	return func(options *LoggerOptions) {
		options.RuntimeInfo = b
	}
}
