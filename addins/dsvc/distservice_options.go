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

// DistServiceOptions 配置分布式服务的节点信息、地址空间及消息处理容量。
type DistServiceOptions struct {
	Version           string            // Version 是对外发布的服务版本。
	Meta              map[string]string // Meta 是对外发布的服务节点元数据。
	DomainRoot        string            // DomainRoot 是消息地址空间的根域。
	RegistrationTTL   time.Duration     // RegistrationTTL 是服务发现注册租约的有效期。
	FutureTimeout     time.Duration     // FutureTimeout 是请求等待响应的默认超时。
	ListenerInboxSize int               // ListenerInboxSize 是每个消息监听器的收件箱容量。
	MsgCreator        gap.IMsgCreator   // MsgCreator 用于按消息 ID 创建解码目标。
}

// With 提供分布式服务 add-in 的 Option 构造方法。
var With _DistServiceOption

type _DistServiceOption struct{}

// Default 返回 svc 根域、30 秒注册租约、5 秒 Future 超时和默认 GAP 消息构建器。
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

// Version 设置对外发布的服务版本。
func (_DistServiceOption) Version(version string) option.Setting[DistServiceOptions] {
	return func(options *DistServiceOptions) {
		options.Version = version
	}
}

// Meta 设置对外发布的服务节点元数据；map 不会复制。
func (_DistServiceOption) Meta(meta map[string]string) option.Setting[DistServiceOptions] {
	return func(options *DistServiceOptions) {
		options.Meta = meta
	}
}

// DomainRoot 设置消息地址空间的根域。
func (_DistServiceOption) DomainRoot(path string) option.Setting[DistServiceOptions] {
	return func(options *DistServiceOptions) {
		options.DomainRoot = path
	}
}

// RegistrationTTL 设置服务发现注册租约有效期，必须不少于三秒。
func (_DistServiceOption) RegistrationTTL(ttl time.Duration) option.Setting[DistServiceOptions] {
	return func(options *DistServiceOptions) {
		if ttl < 3*time.Second {
			exception.Panicf("dsvc: %w: option RegistrationTTL must be >= 3 seconds", core.ErrArgs)
		}
		options.RegistrationTTL = ttl
	}
}

// FutureTimeout 设置请求 Future 的默认超时，必须不少于 300 毫秒。
func (_DistServiceOption) FutureTimeout(d time.Duration) option.Setting[DistServiceOptions] {
	return func(options *DistServiceOptions) {
		if d < 300*time.Millisecond {
			exception.Panicf("dsvc: %w: option FutureTimeout must be >= 0.3 seconds", core.ErrArgs)
		}
		options.FutureTimeout = d
	}
}

// ListenerInboxSize 设置每个消息监听器的收件箱容量，必须大于 0。
func (_DistServiceOption) ListenerInboxSize(size int) option.Setting[DistServiceOptions] {
	return func(options *DistServiceOptions) {
		if size <= 0 {
			exception.Panicf("dsvc: %w: option ListenerInboxSize must be > 0", core.ErrArgs)
		}
		options.ListenerInboxSize = size
	}
}

// MsgCreator 设置 GAP 消息构建器，不得为 nil。
func (_DistServiceOption) MsgCreator(mc gap.IMsgCreator) option.Setting[DistServiceOptions] {
	return func(options *DistServiceOptions) {
		if mc == nil {
			exception.Panicf("dsvc: %w: option MsgCreator can't be assigned to nil", core.ErrArgs)
		}
		options.MsgCreator = mc
	}
}
