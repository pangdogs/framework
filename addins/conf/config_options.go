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

package conf

import (
	"git.golaxy.org/core/utils/option"
	"github.com/spf13/viper"
)

// ConfigOptions 所有选项
type ConfigOptions struct {
	Vipper *viper.Viper
}

var With _ConfigOption

type _ConfigOption struct{}

// Default 默认值
func (_ConfigOption) Default() option.Setting[ConfigOptions] {
	return func(options *ConfigOptions) {
		With.Vipper(nil).Apply(options)
	}
}

func (_ConfigOption) Vipper(v *viper.Viper) option.Setting[ConfigOptions] {
	return func(options *ConfigOptions) {
		options.Vipper = v
	}
}
