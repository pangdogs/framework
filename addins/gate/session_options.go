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

package gate

import (
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/net/gtp/transport"
	"git.golaxy.org/framework/utils/binaryutil"
)

type _SessionOptions struct {
	StateChangedHandler    SessionStateChangedHandler   // 会话状态变化的处理器
	SendDataChan           chan binaryutil.RecycleBytes // 发送数据的channel
	RecvDataChan           chan binaryutil.RecycleBytes // 接收数据的channel
	RecvDataChanRecyclable bool                         // 接收数据的channel中是否使用可回收字节对象
	SendEventChan          chan transport.IEvent        // 发送自定义事件的channel
	RecvEventChan          chan transport.IEvent        // 接收自定义事件的channel
	RecvDataHandler        SessionRecvDataHandler       // 接收数据的处理器（优先级低于监控器）
	RecvEventHandler       SessionRecvEventHandler      // 接收自定义事件的处理器（优先级低于监控器）
}

var sessionWith _SessionOption

type _SessionOption struct{}

func (_SessionOption) Default() option.Setting[_SessionOptions] {
	return func(options *_SessionOptions) {
		sessionWith.StateChangedHandler(nil)(options)
		sessionWith.SendDataChanSize(0)(options)
		sessionWith.RecvDataChanSize(0, false)(options)
		sessionWith.SendEventChanSize(0)(options)
		sessionWith.RecvEventChanSize(0)(options)
		sessionWith.RecvDataHandler(nil)(options)
		sessionWith.RecvEventHandler(nil)(options)
	}
}

func (_SessionOption) StateChangedHandler(handler SessionStateChangedHandler) option.Setting[_SessionOptions] {
	return func(options *_SessionOptions) {
		options.StateChangedHandler = handler
	}
}

func (_SessionOption) SendDataChanSize(size int) option.Setting[_SessionOptions] {
	return func(options *_SessionOptions) {
		if size > 0 {
			options.SendDataChan = make(chan binaryutil.RecycleBytes, size)
		} else {
			options.SendDataChan = nil
		}
	}
}

func (_SessionOption) RecvDataChanSize(size int, recyclable bool) option.Setting[_SessionOptions] {
	return func(options *_SessionOptions) {
		if size > 0 {
			options.RecvDataChan = make(chan binaryutil.RecycleBytes, size)
		} else {
			options.RecvDataChan = nil
		}
		options.RecvDataChanRecyclable = recyclable
	}
}

func (_SessionOption) SendEventChanSize(size int) option.Setting[_SessionOptions] {
	return func(options *_SessionOptions) {
		if size > 0 {
			options.SendEventChan = make(chan transport.IEvent, size)
		} else {
			options.SendEventChan = nil
		}
	}
}

func (_SessionOption) RecvEventChanSize(size int) option.Setting[_SessionOptions] {
	return func(options *_SessionOptions) {
		if size > 0 {
			options.RecvEventChan = make(chan transport.IEvent, size)
		} else {
			options.RecvEventChan = nil
		}
	}
}

func (_SessionOption) RecvDataHandler(handler SessionRecvDataHandler) option.Setting[_SessionOptions] {
	return func(options *_SessionOptions) {
		options.RecvDataHandler = handler
	}
}

func (_SessionOption) RecvEventHandler(handler SessionRecvEventHandler) option.Setting[_SessionOptions] {
	return func(options *_SessionOptions) {
		options.RecvEventHandler = handler
	}
}
