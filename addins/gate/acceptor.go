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
	"context"
	"errors"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/net/gtp/codec"
	"git.golaxy.org/framework/utils/concurrent"
	"net"
)

// _Acceptor 网络连接接受器
type _Acceptor struct {
	gate    *_Gate
	encoder *codec.Encoder
	decoder *codec.Decoder
}

// accept 接受网络连接
func (acc *_Acceptor) accept(conn net.Conn) (*_Session, error) {
	select {
	case <-acc.gate.ctx.Done():
		return nil, errors.New("gate: service shutdown")
	default:
	}

	ctx, _ := context.WithTimeout(acc.gate.ctx, acc.gate.options.AcceptTimeout)

	return acc.handshake(ctx, conn)
}

// newSession 创建会话
func (acc *_Acceptor) newSession(conn net.Conn) (*_Session, error) {
	if conn == nil {
		return nil, errors.New("gate: conn is nil")
	}

	session := &_Session{
		terminated: async.MakeAsyncRet(),
		gate:       acc.gate,
		id:         uid.New(),
		state:      SessionState_Birth,
	}

	session.Context, session.terminate = context.WithCancelCause(acc.gate.ctx)
	session.transceiver.Conn = conn

	// 初始化会话默认选项
	sessionWith.Default()(&session.options)
	sessionWith.SendDataChanSize(acc.gate.options.SessionSendDataChanSize)(&session.options)
	sessionWith.RecvDataChanSize(acc.gate.options.SessionRecvDataChanSize, acc.gate.options.SessionRecvDataChanRecyclable)(&session.options)
	sessionWith.SendEventChanSize(acc.gate.options.SessionSendEventChanSize)(&session.options)
	sessionWith.RecvEventChanSize(acc.gate.options.SessionRecvEventChanSize)(&session.options)

	// 初始化消息事件分发器
	session.eventDispatcher.Transceiver = &session.transceiver
	session.eventDispatcher.RetryTimes = acc.gate.options.IORetryTimes
	session.eventDispatcher.EventHandler = generic.CastDelegate1(session.trans.HandleRecvEvent, session.ctrl.HandleRecvEvent, session.handleRecvEventChan, session.handleRecvEvent)

	// 初始化传输协议
	session.trans.Transceiver = &session.transceiver
	session.trans.RetryTimes = acc.gate.options.IORetryTimes
	session.trans.PayloadHandler = generic.CastDelegate1(session.handleRecvDataChan, session.handleRecvPayload)

	// 初始化控制协议
	session.ctrl.Transceiver = &session.transceiver
	session.ctrl.RetryTimes = acc.gate.options.IORetryTimes
	session.ctrl.HeartbeatHandler = generic.CastDelegate1(session.handleRecvHeartbeat)

	// 初始化监听器
	session.dataWatchers = concurrent.MakeLockedSlice[*_DataWatcher](0, 0)
	session.eventWatchers = concurrent.MakeLockedSlice[*_EventWatcher](0, 0)

	return session, nil
}
