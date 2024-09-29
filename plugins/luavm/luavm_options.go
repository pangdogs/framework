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

package luavm

import (
	"git.golaxy.org/core/utils/option"
)

// LuaVMOptions 所有选项
type LuaVMOptions struct {
	BuiltInLibraries bool
	ShowGoStackTrace bool
}

var With _Option

type _Option struct{}

// Default 默认值
func (_Option) Default() option.Setting[LuaVMOptions] {
	return func(options *LuaVMOptions) {
		With.BuiltInLibraries(true).Apply(options)
		With.ShowGoStackTrace(true).Apply(options)
	}
}

// BuiltInLibraries 使用lua内置库
func (_Option) BuiltInLibraries(b bool) option.Setting[LuaVMOptions] {
	return func(o *LuaVMOptions) {
		o.BuiltInLibraries = b
	}
}

// ShowGoStackTrace 当panic时显示golang的堆栈
func (_Option) ShowGoStackTrace(b bool) option.Setting[LuaVMOptions] {
	return func(o *LuaVMOptions) {
		o.ShowGoStackTrace = b
	}
}
