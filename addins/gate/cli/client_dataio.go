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

	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/transport"
	"git.golaxy.org/framework/utils/concurrent"
	"go.uber.org/zap"
)

type (
	DataHandler = generic.DelegateVoid1[[]byte] // 接收的数据的处理器
)

// IDataIO 数据IO
type IDataIO interface {
	// Send 发送数据
	Send(data []byte) error
	// Listen 监听数据
	Listen(ctx context.Context, handler DataHandler) error
}

type _ClientDataIO struct {
	client    *Client
	listeners concurrent.Listeners[DataHandler, []byte]
}

// Send 发送数据
func (io *_ClientDataIO) Send(data []byte) error {
	return io.client.trans.SendData(data)
}

// Listen 监听数据
func (io *_ClientDataIO) Listen(ctx context.Context, handler DataHandler) error {
	if handler == nil {
		return errors.New("cli: handler is nil")
	}
	return io.addListener(ctx, handler)
}

func (io *_ClientDataIO) init(client *Client) {
	io.client = client
}

func (io *_ClientDataIO) addListener(ctx context.Context, handler DataHandler) error {
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

	listener := io.listeners.Add(handler, io.client.options.DataListenerInboxSize)

	go func() {
		for {
			select {
			case <-ctx.Done():
				io.listeners.Delete(listener)
				io.client.logger.Debug("delete a receive data listener", zap.String("session_id", io.client.SessionId().String()))
				return
			case data := <-listener.Inbox:
				listener.Handler.Call(io.client.options.AutoRecover, io.client.options.ReportError, func(panicError error) bool {
					if panicError != nil {
						io.client.logger.Error("handle receive data panicked", zap.String("session_id", io.client.SessionId().String()), zap.Error(panicError))
					}
					return false
				}, data)
			}
		}
	}()

	io.client.logger.Debug("add a receive data listener", zap.String("session_id", io.client.SessionId().String()))
	return nil
}

func (io *_ClientDataIO) handlePayload(event transport.Event[*gtp.MsgPayload]) {
	rejected := io.listeners.Broadcast(event.Msg.Data)
	if rejected > 0 {
		io.client.logger.Error("some listeners rejected the receive payload due to backpressure",
			zap.String("session_id", io.client.SessionId().String()),
			zap.Uint32("seq", event.Seq),
			zap.Uint32("ack", event.Ack),
			zap.Int("rejected", rejected))
	}
}
