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

package cache_discovery

import (
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/plugins/discovery"
)

// RegistryOptions 所有选项
type RegistryOptions struct {
	Registry discovery.IRegistry // 包装的其他registry插件
}

var With _Option

type _Option struct{}

// Default 默认值
func (_Option) Default() option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		With.Wrap(nil)(options)
	}
}

// Wrap 包装的其他registry插件
func (_Option) Wrap(r discovery.IRegistry) option.Setting[RegistryOptions] {
	return func(o *RegistryOptions) {
		o.Registry = r
	}
}
