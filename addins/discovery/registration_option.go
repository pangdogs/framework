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

package discovery

import (
	"time"

	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/option"
)

// RegisterOptions 所有注册选项
type RegisterOptions struct {
	TTL           time.Duration // 节点TTL
	AutoKeepAlive bool          // 节点自动保活
}

var With _RegisterOption

type _RegisterOption struct{}

// Default 默认注册选项
func (_RegisterOption) Default() option.Setting[RegisterOptions] {
	return func(options *RegisterOptions) {
		With.TTL(30 * time.Second).Apply(options)
		With.AutoKeepAlive(true).Apply(options)
	}
}

// TTL 设置节点TTL
func (_RegisterOption) TTL(ttl time.Duration) option.Setting[RegisterOptions] {
	return func(options *RegisterOptions) {
		if ttl < 3*time.Second {
			exception.Panicf("registry: %w: option TTL can't be set to a value less than 3 seconds", core.ErrArgs)
		}
		options.TTL = ttl
	}
}

// AutoKeepAlive 设置节点自动保活
func (_RegisterOption) AutoKeepAlive(b bool) option.Setting[RegisterOptions] {
	return func(options *RegisterOptions) {
		options.AutoKeepAlive = b
	}
}
