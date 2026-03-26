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

package cli

import (
	"context"
	"errors"

	"git.golaxy.org/framework/net/gtp/transport"
	"git.golaxy.org/framework/utils/concurrent"
	"go.uber.org/zap"
)

type (
	EventHandler = transport.EventHandler // 接收的事件的处理器
)

// IEventIO 事件IO
type IEventIO interface {
	// Send 发送事件
	Send(event transport.IEvent) error
	// Listen 监听事件
	Listen(ctx context.Context, handler EventHandler) error
}

type _ClientEventIO struct {
	client    *Client
	listeners concurrent.Listeners[EventHandler, transport.IEvent]
}

// Send 发送事件
func (io *_ClientEventIO) Send(event transport.IEvent) error {
	return transport.Retry{
		Transceiver: &io.client.transceiver,
		Times:       io.client.options.IORetryTimes,
	}.Send(io.client.transceiver.Send(event))
}

// Listen 监听事件
func (io *_ClientEventIO) Listen(ctx context.Context, handler EventHandler) error {
	if handler == nil {
		return errors.New("cli: handler is nil")
	}
	return io.addListener(ctx, handler)
}

func (io *_ClientEventIO) init(client *Client) {
	io.client = client
}

func (io *_ClientEventIO) addListener(ctx context.Context, handler EventHandler) error {
	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case <-io.client.closed:
		return errors.New("cli: client closed")
	default:
	}

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-ctx.Done():
		case <-io.client.closed:
		}
		cancel()
	}()

	listener := io.listeners.Add(handler, io.client.options.EventListenerInboxSize)

	go func() {
		for {
			select {
			case <-ctx.Done():
				io.listeners.Delete(listener)
				io.client.logger.Debug("delete a receive event listener", zap.String("session_id", io.client.SessionId().String()))
				return
			case event := <-listener.Inbox:
				listener.Handler.Call(io.client.options.AutoRecover, io.client.options.ReportError, func(panicError error) bool {
					if panicError != nil {
						io.client.logger.Error("handle receive event panicked", zap.String("session_id", io.client.SessionId().String()), zap.Error(panicError))
					}
					return false
				}, event)
			}
		}
	}()

	io.client.logger.Debug("add a receive event listener", zap.String("session_id", io.client.SessionId().String()))
	return nil
}

func (io *_ClientEventIO) handleEvent(event transport.IEvent) {
	rejected := io.listeners.Broadcast(event)
	if rejected > 0 {
		io.client.logger.Error("some listeners rejected the receive event due to backpressure",
			zap.String("session_id", io.client.SessionId().String()),
			zap.Uint32("seq", event.Seq),
			zap.Uint32("ack", event.Ack),
			zap.Uint8("msg_id", event.Msg.MsgId()),
			zap.Int("rejected", rejected))
	}
}
