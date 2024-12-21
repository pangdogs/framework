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
	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/net/gap"
	"time"
)

type (
	RecvMsgHandler = generic.Delegate2[string, gap.MsgPacket, error] // 接收消息的处理器
)

// DistServiceOptions 所有选项
type DistServiceOptions struct {
	Version           string            // 服务版本号
	Meta              map[string]string // 服务元数据，以键值对的形式保存附加信息
	DomainRoot        string            // 服务地址根域
	TTL               time.Duration     // 服务信息TTL
	RefreshTTL        bool              // 主动刷新服务信息TTL
	FutureTimeout     time.Duration     // 异步模型Future超时时间
	DecoderMsgCreator gap.IMsgCreator   // 消息包解码器的消息构建器
	RecvMsgHandler    RecvMsgHandler    // 接收消息的处理器（优先级低于监控器）
}

var With _Option

type _Option struct{}

// Default 默认值
func (_Option) Default() option.Setting[DistServiceOptions] {
	return func(options *DistServiceOptions) {
		With.Version("")(options)
		With.Meta(nil)(options)
		With.DomainRoot("svc")(options)
		With.TTL(0, false)(options)
		With.FutureTimeout(5 * time.Second)(options)
		With.DecoderMsgCreator(gap.DefaultMsgCreator())(options)
		With.RecvMsgHandler(nil)(options)
	}
}

// Version 服务版本号
func (_Option) Version(version string) option.Setting[DistServiceOptions] {
	return func(o *DistServiceOptions) {
		o.Version = version
	}
}

// Meta 服务元数据，以键值对的形式保存附加信息
func (_Option) Meta(meta map[string]string) option.Setting[DistServiceOptions] {
	return func(o *DistServiceOptions) {
		o.Meta = meta
	}
}

// DomainRoot 服务地址根域
func (_Option) DomainRoot(path string) option.Setting[DistServiceOptions] {
	return func(o *DistServiceOptions) {
		o.DomainRoot = path
	}
}

// TTL 服务信息TTL
func (_Option) TTL(ttl time.Duration, refresh bool) option.Setting[DistServiceOptions] {
	return func(o *DistServiceOptions) {
		o.TTL = ttl
		o.RefreshTTL = refresh
	}
}

// FutureTimeout 异步模型Future超时时间
func (_Option) FutureTimeout(d time.Duration) option.Setting[DistServiceOptions] {
	return func(options *DistServiceOptions) {
		if d <= 0 {
			exception.Panicf("%w: option FutureTimeout can't be set to a value less equal 0", core.ErrArgs)
		}
		options.FutureTimeout = d
	}
}

// DecoderMsgCreator 消息包解码器的消息构建器
func (_Option) DecoderMsgCreator(mc gap.IMsgCreator) option.Setting[DistServiceOptions] {
	return func(options *DistServiceOptions) {
		if mc == nil {
			exception.Panicf("%w: option DecoderMsgCreator can't be assigned to nil", core.ErrArgs)
		}
		options.DecoderMsgCreator = mc
	}
}

// RecvMsgHandler 接收消息的处理器
func (_Option) RecvMsgHandler(handler RecvMsgHandler) option.Setting[DistServiceOptions] {
	return func(options *DistServiceOptions) {
		options.RecvMsgHandler = handler
	}
}
