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

package dsvc

import (
	"time"

	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/net/gap"
)

// DistServiceOptions 所有选项
type DistServiceOptions struct {
	Version           string            // 服务版本号
	Meta              map[string]string // 服务元数据，以键值对的形式保存附加信息
	DomainRoot        string            // 服务地址根域
	RegistrationTTL   time.Duration     // 服务注册信息TTL
	FutureTimeout     time.Duration     // 异步模型Future超时时间
	ListenerInboxSize int               // 消息监听器的inbox缓存大小
	MsgCreator        gap.IMsgCreator   // 消息包解码器的消息构建器
}

var With _DistServiceOption

type _DistServiceOption struct{}

// Default 默认值
func (_DistServiceOption) Default() option.Setting[DistServiceOptions] {
	return func(options *DistServiceOptions) {
		With.Version("").Apply(options)
		With.Meta(nil).Apply(options)
		With.DomainRoot("svc").Apply(options)
		With.RegistrationTTL(30 * time.Second).Apply(options)
		With.FutureTimeout(5 * time.Second).Apply(options)
		With.ListenerInboxSize(256 * 1024).Apply(options)
		With.MsgCreator(gap.DefaultMsgCreator()).Apply(options)
	}
}

// Version 服务版本号
func (_DistServiceOption) Version(version string) option.Setting[DistServiceOptions] {
	return func(options *DistServiceOptions) {
		options.Version = version
	}
}

// Meta 服务元数据，以键值对的形式保存附加信息
func (_DistServiceOption) Meta(meta map[string]string) option.Setting[DistServiceOptions] {
	return func(options *DistServiceOptions) {
		options.Meta = meta
	}
}

// DomainRoot 服务地址根域
func (_DistServiceOption) DomainRoot(path string) option.Setting[DistServiceOptions] {
	return func(options *DistServiceOptions) {
		options.DomainRoot = path
	}
}

// RegistrationTTL 服务注册信息TTL
func (_DistServiceOption) RegistrationTTL(ttl time.Duration) option.Setting[DistServiceOptions] {
	return func(options *DistServiceOptions) {
		if ttl < 3*time.Second {
			exception.Panicf("dsvc: %w: option RegistrationTTL must be >= 3 seconds", core.ErrArgs)
		}
		options.RegistrationTTL = ttl
	}
}

// FutureTimeout 异步模型Future超时时间
func (_DistServiceOption) FutureTimeout(d time.Duration) option.Setting[DistServiceOptions] {
	return func(options *DistServiceOptions) {
		if d < 300*time.Millisecond {
			exception.Panicf("dsvc: %w: option FutureTimeout must be >= 0.3 seconds", core.ErrArgs)
		}
		options.FutureTimeout = d
	}
}

// ListenerInboxSize 消息监听器的inbox缓存大小
func (_DistServiceOption) ListenerInboxSize(size int) option.Setting[DistServiceOptions] {
	return func(options *DistServiceOptions) {
		if size <= 0 {
			exception.Panicf("dsvc: %w: option ListenerInboxSize must be > 0", core.ErrArgs)
		}
		options.ListenerInboxSize = size
	}
}

// MsgCreator 消息包解码器的消息构建器
func (_DistServiceOption) MsgCreator(mc gap.IMsgCreator) option.Setting[DistServiceOptions] {
	return func(options *DistServiceOptions) {
		if mc == nil {
			exception.Panicf("dsvc: %w: option MsgCreator can't be assigned to nil", core.ErrArgs)
		}
		options.MsgCreator = mc
	}
}
