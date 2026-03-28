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

	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/transport"
	"git.golaxy.org/framework/utils/binaryutil"
	"git.golaxy.org/framework/utils/concurrent"
	"go.uber.org/zap"
)

type (
	DataHandler  = generic.DelegateVoid1[[]byte]
	EventHandler = transport.EventHandler
)

// IDataIO 数据IO
type IDataIO interface {
	// Send 发送数据
	Send(data []byte) error
	// Listen 监听数据
	Listen(ctx context.Context, handler DataHandler) error
}

// IEventIO 事件IO
type IEventIO interface {
	// Send 发送事件
	Send(event transport.IEvent) error
	// Listen 监听事件
	Listen(ctx context.Context, handler EventHandler) error
}

type _ClientIO struct {
	client         *Client
	barrier        generic.Barrier
	terminated     async.FutureVoid
	dataChan       *generic.UnboundedChannel[binaryutil.Bytes]
	eventChan      *generic.UnboundedChannel[transport.IEvent]
	dataListeners  concurrent.Listeners[DataHandler, []byte]
	eventListeners concurrent.Listeners[EventHandler, transport.IEvent]
}

func (io *_ClientIO) init(client *Client) {
	io.client = client
	io.terminated = async.NewFutureVoid()
	io.dataChan = generic.NewUnboundedChannel[binaryutil.Bytes]()
	io.eventChan = generic.NewUnboundedChannel[transport.IEvent]()
}

func (io *_ClientIO) sendLoop() {
loop:
	for {
		select {
		case <-io.client.Done():
			break loop

		case buff := <-io.dataChan.Out():
			if err := io.client.trans.SendData(buff.Payload()); err != nil {
				io.client.logger.Error("client send data failed",
					zap.String("session_id", io.client.SessionId().String()),
					zap.Int64("migrations", io.client.Migrations()),
					zap.Error(err))
			}
			buff.Release()

		case event := <-io.eventChan.Out():
			err := transport.Retry{
				Transceiver: &io.client.transceiver,
				Times:       io.client.options.IORetryTimes,
			}.Send(io.client.transceiver.Send(event))
			if err != nil {
				io.client.logger.Error("client send event failed",
					zap.String("session_id", io.client.SessionId().String()),
					zap.Int64("migrations", io.client.Migrations()),
					zap.Error(err))
			}
		}
	}

	io.barrier.Close()
	io.barrier.Wait()

	io.dataChan.Close()
	io.eventChan.Close()

	for buff := range io.dataChan.Out() {
		if err := io.client.trans.SendData(buff.Payload()); err != nil {
			io.client.logger.Error("client send data failed",
				zap.String("session_id", io.client.SessionId().String()),
				zap.Int64("migrations", io.client.Migrations()),
				zap.Error(err))
		}
		buff.Release()
	}

	for event := range io.eventChan.Out() {
		err := transport.Retry{
			Transceiver: &io.client.transceiver,
			Times:       io.client.options.IORetryTimes,
		}.Send(io.client.transceiver.Send(event))
		if err != nil {
			io.client.logger.Error("client send event failed",
				zap.String("session_id", io.client.SessionId().String()),
				zap.Int64("migrations", io.client.Migrations()),
				zap.Error(err))
		}
	}

	async.ReturnVoid(io.terminated)
}

func (io *_ClientIO) handlePayload(event transport.Event[*gtp.MsgPayload]) {
	rejected := io.dataListeners.Broadcast(event.Msg.Data)
	if rejected > 0 {
		io.client.logger.Error("some listeners rejected the receive payload due to backpressure",
			zap.String("session_id", io.client.SessionId().String()),
			zap.Uint32("seq", event.Seq),
			zap.Uint32("ack", event.Ack),
			zap.Int("rejected", rejected))
	}
}

func (io *_ClientIO) handleEvent(event transport.IEvent) {
	rejected := io.eventListeners.Broadcast(event)
	if rejected > 0 {
		io.client.logger.Error("some listeners rejected the receive event due to backpressure",
			zap.String("session_id", io.client.SessionId().String()),
			zap.Uint32("seq", event.Seq),
			zap.Uint32("ack", event.Ack),
			zap.Uint8("msg_id", event.Msg.MsgId()),
			zap.Int("rejected", rejected))
	}
}

type _ClientDataIO _ClientIO

func (io *_ClientDataIO) Send(data []byte) error {
	if !io.barrier.Join(1) {
		return errors.New("cli: client data i/o is terminating")
	}
	defer io.barrier.Done()

	io.dataChan.In() <- binaryutil.CloneBytes(true, data)
	return nil
}

func (io *_ClientDataIO) Listen(ctx context.Context, handler DataHandler) error {
	if handler == nil {
		return errors.New("cli: handler is nil")
	}
	return io.addListener(ctx, handler)
}

func (io *_ClientDataIO) addListener(ctx context.Context, handler DataHandler) error {
	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case <-io.client.Done():
		return errors.New("cli: client data i/o is terminating")
	default:
	}

	if !io.barrier.Join(1) {
		return errors.New("cli: client data i/o is terminating")
	}

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-ctx.Done():
		case <-io.client.Done():
		}
		cancel()
	}()

	listener := io.dataListeners.Add(handler, io.client.options.DataListenerInboxSize)

	go func() {
		defer io.barrier.Done()
		for {
			select {
			case <-ctx.Done():
				io.dataListeners.Delete(listener)
				io.client.logger.Debug("delete a receive data listener", zap.String("session_id", io.client.SessionId().String()))
				return
			case data := <-listener.Inbox:
				listener.Handler.Call(io.client.options.AutoRecover, io.client.options.ReportError, func(panicError error) bool {
					if panicError != nil {
						io.client.logger.Error("handle receive data panicked",
							zap.String("session_id", io.client.SessionId().String()),
							zap.Error(panicError))
					}
					return false
				}, data)
			}
		}
	}()

	io.client.logger.Debug("add a receive data listener", zap.String("session_id", io.client.SessionId().String()))
	return nil
}

type _ClientEventIO _ClientIO

func (io *_ClientEventIO) Send(event transport.IEvent) error {
	if !io.barrier.Join(1) {
		return errors.New("cli: client event i/o is terminating")
	}
	defer io.barrier.Done()

	io.eventChan.In() <- event
	return nil
}

func (io *_ClientEventIO) Listen(ctx context.Context, handler EventHandler) error {
	if handler == nil {
		return errors.New("cli: handler is nil")
	}
	return io.addListener(ctx, handler)
}

func (io *_ClientEventIO) addListener(ctx context.Context, handler EventHandler) error {
	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case <-io.client.Done():
		return errors.New("cli: client event i/o is terminating")
	default:
	}

	if !io.barrier.Join(1) {
		return errors.New("cli: client event i/o is terminating")
	}

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-ctx.Done():
		case <-io.client.Done():
		}
		cancel()
	}()

	listener := io.eventListeners.Add(handler, io.client.options.EventListenerInboxSize)

	go func() {
		defer io.barrier.Done()
		for {
			select {
			case <-ctx.Done():
				io.eventListeners.Delete(listener)
				io.client.logger.Debug("delete a receive event listener", zap.String("session_id", io.client.SessionId().String()))
				return
			case event := <-listener.Inbox:
				listener.Handler.Call(io.client.options.AutoRecover, io.client.options.ReportError, func(panicError error) bool {
					if panicError != nil {
						io.client.logger.Error("handle receive event panicked",
							zap.String("session_id", io.client.SessionId().String()),
							zap.Error(panicError))
					}
					return false
				}, event)
			}
		}
	}()

	io.client.logger.Debug("add a receive event listener", zap.String("session_id", io.client.SessionId().String()))
	return nil
}
