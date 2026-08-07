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

// _Acceptor 接受底层连接并按 gate 配置完成服务端 GTP 握手。
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

// genSessionId 生成新的会话 ID。
func (acc *_Acceptor) genSessionId() uid.Id {
	return uid.New()
}

// newSession 创建尚未绑定连接的会话，并装配传输、控制和 I/O 处理器。
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

	// 分发器按传输、控制、业务事件的顺序尝试处理同一事件。
	session.eventDispatcher.AutoRecover = acc.svcCtx.AutoRecover()
	session.eventDispatcher.ReportError = acc.svcCtx.ReportError()
	session.eventDispatcher.Transceiver = &session.transceiver
	session.eventDispatcher.RetryTimes = acc.options.IORetryTimes
	session.eventDispatcher.EventHandler = generic.CastDelegateVoid1(session.trans.HandleEvent, session.ctrl.HandleEvent, session.io.handleEvent)

	// 传输协议负责业务负载与重试。
	session.trans.AutoRecover = acc.svcCtx.AutoRecover()
	session.trans.ReportError = acc.svcCtx.ReportError()
	session.trans.Transceiver = &session.transceiver
	session.trans.RetryTimes = acc.options.IORetryTimes
	session.trans.PayloadHandler = generic.CastDelegateVoid1(session.io.handlePayload)

	// 控制协议负责心跳和链路重置等会话控制事件。
	session.ctrl.AutoRecover = acc.svcCtx.AutoRecover()
	session.ctrl.ReportError = acc.svcCtx.ReportError()
	session.ctrl.Transceiver = &session.transceiver
	session.ctrl.RetryTimes = acc.options.IORetryTimes
	session.ctrl.HeartbeatHandler = generic.CastDelegateVoid1(session.handleHeartbeat)

	// I/O 门面使用独立队列异步发送数据和事件。
	session.io.init(session)

	return session
}
