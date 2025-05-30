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
	"github.com/spf13/pflag"
	"time"
)

// ConfigOptions 所有选项
type ConfigOptions struct {
	Defaults                             map[string]any // 默认配置
	Flags                                *pflag.FlagSet // 启动命令参数
	AutomaticEnv                         bool           // 合并环境变量
	EnvPrefix                            string         // 环境变量前缀
	LocalPath                            string         // 本地配置文件路径
	RemoteProvider                       string         // 远端配置类型（etcd3,consul...）
	RemoteEndpoint                       string         // 远端地址
	RemotePath                           string         // 远端路径
	AutoHotFix                           bool           // 自动热更新
	AutoHotFixRemoteCheckingIntervalTime time.Duration  // 自动热更新远端配置检测间隔时间
}

var With _Option

type _Option struct{}

// Default 默认值
func (_Option) Default() option.Setting[ConfigOptions] {
	return func(options *ConfigOptions) {
		With.Defaults(nil).Apply(options)
		With.Flags(nil).Apply(options)
		With.AutomaticEnv(false).Apply(options)
		With.EnvPrefix("").Apply(options)
		With.Local("").Apply(options)
		With.Remote("", "", "").Apply(options)
		With.AutoHotFix(false).Apply(options)
		With.AutoHotFixRemoteCheckingIntervalTime(time.Minute).Apply(options)
	}
}

// Defaults 默认配置
func (_Option) Defaults(dict map[string]any) option.Setting[ConfigOptions] {
	return func(options *ConfigOptions) {
		options.Defaults = dict
	}
}

// AutomaticEnv 合并环境变量
func (_Option) AutomaticEnv(b bool) option.Setting[ConfigOptions] {
	return func(options *ConfigOptions) {
		options.AutomaticEnv = b
	}
}

// EnvPrefix 环境变量前缀
func (_Option) EnvPrefix(prefix string) option.Setting[ConfigOptions] {
	return func(options *ConfigOptions) {
		options.EnvPrefix = prefix
	}
}

// Flags 启动命令参数
func (_Option) Flags(flags *pflag.FlagSet) option.Setting[ConfigOptions] {
	return func(options *ConfigOptions) {
		options.Flags = flags
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

// AutoHotFixRemoteCheckingIntervalTime 自动热更新远端配置检测间隔时间
func (_Option) AutoHotFixRemoteCheckingIntervalTime(d time.Duration) option.Setting[ConfigOptions] {
	return func(options *ConfigOptions) {
		options.AutoHotFixRemoteCheckingIntervalTime = d
	}
}
