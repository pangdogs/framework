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
	"git.golaxy.org/framework/net/gtp/transport"
	"git.golaxy.org/framework/utils/concurrent"
	"go.uber.org/zap"
)

type (
	SessionEventHandler = generic.DelegateVoid2[ISession, transport.IEvent] // 会话接收的事件的处理器
)

// IEventIO 事件IO
type IEventIO interface {
	// Send 发送事件
	Send(event transport.IEvent) error
	// Listen 监听事件
	Listen(ctx context.Context, handler SessionEventHandler) error
}

type _SessionEventIO struct {
	session   *_Session
	listeners concurrent.Listeners[SessionEventHandler, transport.IEvent]
}

// Send 发送事件
func (io *_SessionEventIO) Send(event transport.IEvent) error {
	return transport.Retry{
		Transceiver: &io.session.transceiver,
		Times:       io.session.gate.options.IORetryTimes,
	}.Send(io.session.transceiver.Send(event))
}

// Listen 监听事件
func (io *_SessionEventIO) Listen(ctx context.Context, handler SessionEventHandler) error {
	if handler == nil {
		return errors.New("gate: handler is nil")
	}
	return io.addListener(ctx, handler)
}

func (io *_SessionEventIO) init(session *_Session) {
	io.session = session
}

func (io *_SessionEventIO) addListener(ctx context.Context, handler SessionEventHandler) error {
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

	listener := io.listeners.Add(handler, io.session.gate.options.SessionEventListenerInboxSize)

	go func() {
		for {
			select {
			case <-ctx.Done():
				io.listeners.Delete(listener)
				log.L(io.session.gate.svcCtx).Debug("delete a session event listener", zap.String("session_id", io.session.Id().String()))
				return
			case event := <-listener.Inbox:
				listener.Handler.Call(io.session.gate.svcCtx.AutoRecover(), io.session.gate.svcCtx.ReportError(), func(panicError error) bool {
					if panicError != nil {
						log.L(io.session.gate.svcCtx).Error("handle session event panicked",
							zap.String("session_id", io.session.Id().String()),
							zap.Error(panicError))
					}
					return false
				}, io.session, event)
			}
		}
	}()

	log.L(io.session.gate.svcCtx).Debug("add a session event listener", zap.String("session_id", io.session.Id().String()))
	return nil
}

func (io *_SessionEventIO) handleEvent(event transport.IEvent) {
	rejected := io.listeners.Broadcast(event)
	if rejected > 0 {
		log.L(io.session.gate.svcCtx).Error("some listeners rejected the receive event due to backpressure",
			zap.String("session_id", io.session.Id().String()),
			zap.Uint32("seq", event.Seq),
			zap.Uint32("ack", event.Ack),
			zap.Uint8("msg_id", event.Msg.MsgId()),
			zap.Int("rejected", rejected))
	}
}
