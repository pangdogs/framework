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
	Format         string         // 配置格式（json,yaml,ini...）
	LocalPath      string         // 本地配置文件路径
	RemoteProvider string         // 远端配置类型（etcd3,consul...）
	RemoteEndpoint string         // 远端地址
	RemotePath     string         // 远端路径
	AutoHotFix     bool           // 自动热更新
	Defaults       map[string]any // 默认配置
	MergeEnv       bool           // 合并环境变量
	MergeConf      *viper.Viper   // 合并配置
}

var With _Option

type _Option struct{}

// Default 默认值
func (_Option) Default() option.Setting[ConfigOptions] {
	return func(options *ConfigOptions) {
		With.Format("json").Apply(options)
		With.Local("").Apply(options)
		With.Remote("", "", "").Apply(options)
		With.AutoHotFix(false).Apply(options)
		With.Defaults(nil).Apply(options)
		With.MergeEnv(false).Apply(options)
		With.MergeConf(nil).Apply(options)
	}
}

// Format 配置格式（json,yaml,ini...）
func (_Option) Format(format string) option.Setting[ConfigOptions] {
	return func(options *ConfigOptions) {
		options.Format = format
	}
}

// Local 本地配置
func (_Option) Local(path string) option.Setting[ConfigOptions] {
	return func(options *ConfigOptions) {
		options.LocalPath = path
	}
}

// Remote 远端配置
func (_Option) Remote(provider, endpoint, path string) option.Setting[ConfigOptions] {
	return func(options *ConfigOptions) {
		options.RemoteProvider = provider
		options.RemoteEndpoint = endpoint
		options.RemotePath = path
	}
}

// AutoHotFix 是否热更新
func (_Option) AutoHotFix(b bool) option.Setting[ConfigOptions] {
	return func(options *ConfigOptions) {
		options.AutoHotFix = b
	}
}

// Defaults 默认配置
func (_Option) Defaults(dict map[string]any) option.Setting[ConfigOptions] {
	return func(options *ConfigOptions) {
		options.Defaults = dict
	}
}

// MergeEnv 合并环境变量
func (_Option) MergeEnv(b bool) option.Setting[ConfigOptions] {
	return func(options *ConfigOptions) {
		options.MergeEnv = b
	}
}

// MergeConf 合并配置
func (_Option) MergeConf(conf *viper.Viper) option.Setting[ConfigOptions] {
	return func(options *ConfigOptions) {
		options.MergeConf = conf
	}
}
