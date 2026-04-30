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
	"net"

	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/net/gtp/codec"
)

// _Acceptor 网络连接接受器
type _Acceptor struct {
	*_Gate
	encoder *codec.Encoder
	decoder *codec.Decoder
}

// accept 接受网络连接并完成握手。
// 返回的 bool 表示该连接是否为旧会话迁移连接。
func (acc *_Acceptor) accept(conn net.Conn) (*_Session, bool, error) {
	select {
	case <-acc.ctx.Done():
		return nil, false, errors.New("gate: service shutdown")
	default:
	}
	ctx, cancel := context.WithTimeout(acc.ctx, acc.options.AcceptTimeout)
	defer cancel()

	return acc.handshake(ctx, conn)
}

// genSessionId 生成会话ID
func (acc *_Acceptor) genSessionId() uid.Id {
	return uid.New()
}

// newSession 创建会话，无连接初始状态
func (acc *_Acceptor) newSession(id uid.Id, userId, token string, extensions []byte) *_Session {
	session := &_Session{
		closed:        async.NewFutureVoid(),
		gate:          acc._Gate,
		id:            id,
		userId:        userId,
		token:         token,
		extensions:    extensions,
		migrationChan: make(chan struct{}),
	}
	session.Context, session.close = context.WithCancelCause(acc.ctx)

	// 初始化消息事件分发器
	session.eventDispatcher.AutoRecover = acc.svcCtx.AutoRecover()
	session.eventDispatcher.ReportError = acc.svcCtx.ReportError()
	session.eventDispatcher.Transceiver = &session.transceiver
	session.eventDispatcher.RetryTimes = acc.options.IORetryTimes
	session.eventDispatcher.EventHandler = generic.CastDelegateVoid1(session.trans.HandleEvent, session.ctrl.HandleEvent, session.io.handleEvent)

	// 初始化传输协议
	session.trans.AutoRecover = acc.svcCtx.AutoRecover()
	session.trans.ReportError = acc.svcCtx.ReportError()
	session.trans.Transceiver = &session.transceiver
	session.trans.RetryTimes = acc.options.IORetryTimes
	session.trans.PayloadHandler = generic.CastDelegateVoid1(session.io.handlePayload)

	// 初始化控制协议
	session.ctrl.AutoRecover = acc.svcCtx.AutoRecover()
	session.ctrl.ReportError = acc.svcCtx.ReportError()
	session.ctrl.Transceiver = &session.transceiver
	session.ctrl.RetryTimes = acc.options.IORetryTimes
	session.ctrl.HeartbeatHandler = generic.CastDelegateVoid1(session.handleHeartbeat)

	// 初始化IO
	session.io.init(session)

	return session
}
