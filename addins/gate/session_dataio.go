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

	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/transport"
	"git.golaxy.org/framework/utils/concurrent"
	"go.uber.org/zap"
)

type (
	SessionDataHandler = generic.DelegateVoid2[ISession, []byte] // 会话接收的数据的处理器
)

// IDataIO 数据IO
type IDataIO interface {
	// Send 发送数据
	Send(data []byte) error
	// Listen 监听数据
	Listen(ctx context.Context, handler SessionDataHandler) error
}

type _SessionDataIO struct {
	session   *_Session
	listeners concurrent.Listeners[SessionDataHandler, []byte]
}

// Send 发送数据
func (io *_SessionDataIO) Send(data []byte) error {
	return io.session.trans.SendData(data)
}

// Listen 监听数据
func (io *_SessionDataIO) Listen(ctx context.Context, handler SessionDataHandler) error {
	if handler == nil {
		return errors.New("gate: handler is nil")
	}
	return io.addListener(ctx, handler)
}

func (io *_SessionDataIO) init(session *_Session) {
	io.session = session
}

func (io *_SessionDataIO) addListener(ctx context.Context, handler SessionDataHandler) error {
	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case <-io.session.closed:
		return errors.New("gate: session closed")
	default:
	}

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-ctx.Done():
		case <-io.session.closed:
		}
		cancel()
	}()

	listener := io.listeners.Add(handler, io.session.gate.options.SessionDataListenerInboxSize)

	go func() {
		for {
			select {
			case <-ctx.Done():
				io.listeners.Delete(listener)
				log.L(io.session.gate.svcCtx).Debug("delete a session data listener", zap.String("session_id", io.session.Id().String()))
				return
			case data := <-listener.Inbox:
				listener.Handler.Call(io.session.gate.svcCtx.AutoRecover(), io.session.gate.svcCtx.ReportError(), func(panicError error) bool {
					if panicError != nil {
						log.L(io.session.gate.svcCtx).Error("handle session data panicked",
							zap.String("session_id", io.session.Id().String()),
							zap.Error(panicError))
					}
					return false
				}, io.session, data)
			}
		}
	}()

	log.L(io.session.gate.svcCtx).Debug("add a session data listener", zap.String("session_id", io.session.Id().String()))
	return nil
}

func (io *_SessionDataIO) handlePayload(event transport.Event[*gtp.MsgPayload]) {
	rejected := io.listeners.Broadcast(event.Msg.Data)
	if rejected > 0 {
		log.L(io.session.gate.svcCtx).Error("some listeners rejected the receive payload due to backpressure",
			zap.String("session_id", io.session.Id().String()),
			zap.Uint32("seq", event.Seq),
			zap.Uint32("ack", event.Ack),
			zap.Int("rejected", rejected))
	}
}
