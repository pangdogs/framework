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
	"errors"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/option"
)

// SessionSettings 会话设置
type SessionSettings struct {
	session                    *_Session
	settings                   []option.Setting[_SessionOptions]
	CurrStateChangedHandler    SessionStateChangedHandler
	CurrSendDataChanSize       int
	CurrRecvDataChanSize       int
	CurrRecvDataChanRecyclable bool
	CurrSendEventChanSize      int
	CurrRecvEventChanSize      int
	CurrRecvDataHandler        SessionRecvDataHandler
	CurrRecvEventHandler       SessionRecvEventHandler
}

// StateChangedHandler 设置会话状态变化的处理器
func (s SessionSettings) StateChangedHandler(handler SessionStateChangedHandler) SessionSettings {
	s.settings = append(s.settings, sessionWith.StateChangedHandler(handler))
	return s
}

// SendDataChanSize 设置发送数据的channel的大小，<=0表示不使用channel
func (s SessionSettings) SendDataChanSize(size int) SessionSettings {
	s.settings = append(s.settings, sessionWith.SendDataChanSize(size))
	return s
}

// RecvDataChanSize 设置接收数据的channel的大小，<=0表示不使用channel
func (s SessionSettings) RecvDataChanSize(size int, recyclable bool) SessionSettings {
	s.settings = append(s.settings, sessionWith.RecvDataChanSize(size, recyclable))
	return s
}

// SendEventChanSize 设置发送自定义事件的channel的大小，<=0表示不使用channel
func (s SessionSettings) SendEventChanSize(size int) SessionSettings {
	s.settings = append(s.settings, sessionWith.SendEventChanSize(size))
	return s
}

// RecvEventChanSize 设置接收自定义事件的channel的大小，<=0表示不使用channel
func (s SessionSettings) RecvEventChanSize(size int) SessionSettings {
	s.settings = append(s.settings, sessionWith.RecvEventChanSize(size))
	return s
}

// RecvDataHandler 设置接收的数据的处理器
func (s SessionSettings) RecvDataHandler(handler SessionRecvDataHandler) SessionSettings {
	s.settings = append(s.settings, sessionWith.RecvDataHandler(handler))
	return s
}

// RecvEventHandler 设置接收自定义事件的处理器
func (s SessionSettings) RecvEventHandler(handler SessionRecvEventHandler) SessionSettings {
	s.settings = append(s.settings, sessionWith.RecvEventHandler(handler))
	return s
}

// Change 执行修改
func (s SessionSettings) Change() error {
	if s.session == nil {
		exception.Panic("session is nil")
	}

	s.session.Lock()
	defer s.session.Unlock()

	switch s.session.state {
	case SessionState_Birth, SessionState_Handshake, SessionState_Confirmed:
		break
	default:
		return errors.New("incorrect session state")
	}

	option.Change(&s.session.options, s.settings...)

	return nil
}
