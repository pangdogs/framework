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

package goscr

import (
	"git.golaxy.org/core/utils/option"
	"reflect"
)

// GoScrOptions 所有选项
type GoScrOptions struct {
	PathList    []string
	SymbolsList []map[string]map[string]reflect.Value
	AutoHotFix  bool
}

var With _Option

type _Option struct{}

// Default 默认值
func (_Option) Default() option.Setting[GoScrOptions] {
	return func(options *GoScrOptions) {
		With.PathList()(options)
		With.SymbolsList()(options)
		With.AutoHotFix(true)(options)
	}
}

func (_Option) PathList(l ...string) option.Setting[GoScrOptions] {
	return func(options *GoScrOptions) {
		options.PathList = l
	}
}

func (_Option) SymbolsList(l ...map[string]map[string]reflect.Value) option.Setting[GoScrOptions] {
	return func(options *GoScrOptions) {
		options.SymbolsList = l
	}
}

func (_Option) AutoHotFix(b bool) option.Setting[GoScrOptions] {
	return func(options *GoScrOptions) {
		options.AutoHotFix = b
	}
}
