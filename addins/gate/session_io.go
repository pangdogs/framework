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
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/transport"
	"git.golaxy.org/framework/utils/binaryutil"
	"git.golaxy.org/framework/utils/concurrent"
	"go.uber.org/zap"
)

type (
	SessionDataHandler  = generic.DelegateVoid2[ISession, []byte]           // 会话接收的数据的处理器
	SessionEventHandler = generic.DelegateVoid2[ISession, transport.IEvent] // 会话接收的事件的处理器
)

// IDataIO 数据IO
type IDataIO interface {
	// Send 发送数据
	Send(data []byte) error
	// Listen 监听数据
	Listen(ctx context.Context, handler SessionDataHandler) error
}

// IEventIO 事件IO
type IEventIO interface {
	// Send 发送事件
	Send(event transport.IEvent) error
	// Listen 监听事件
	Listen(ctx context.Context, handler SessionEventHandler) error
}

type _SessionIO struct {
	session        *_Session
	barrier        generic.Barrier
	terminated     async.FutureVoid
	dataChan       *generic.UnboundedChannel[binaryutil.Bytes]
	eventChan      *generic.UnboundedChannel[transport.IEvent]
	dataListeners  concurrent.Listeners[SessionDataHandler, []byte]
	eventListeners concurrent.Listeners[SessionEventHandler, transport.IEvent]
}

func (io *_SessionIO) init(session *_Session) {
	io.session = session
	io.terminated = async.NewFutureVoid()
	io.dataChan = generic.NewUnboundedChannel[binaryutil.Bytes]()
	io.eventChan = generic.NewUnboundedChannel[transport.IEvent]()
}

func (io *_SessionIO) sendLoop() {
loop:
	for {
		select {
		case <-io.session.Done():
			break loop

		case buff := <-io.dataChan.Out():
			if err := io.session.trans.SendData(buff.Payload()); err != nil {
				log.L(io.session.gate.svcCtx).Error("session send data error",
					zap.String("session_id", io.session.Id().String()),
					zap.String("local", io.session.NetAddr().Local.String()),
					zap.String("remote", io.session.NetAddr().Remote.String()),
					zap.Int64("migrations", io.session.Migrations()),
					zap.Error(err))
			}
			buff.Release()

		case event := <-io.eventChan.Out():
			err := transport.Retry{
				Transceiver: &io.session.transceiver,
				Times:       io.session.gate.options.IORetryTimes,
			}.Send(io.session.transceiver.Send(event))
			if err != nil {
				log.L(io.session.gate.svcCtx).Error("session send event failed",
					zap.String("session_id", io.session.Id().String()),
					zap.String("local", io.session.NetAddr().Local.String()),
					zap.String("remote", io.session.NetAddr().Remote.String()),
					zap.Int64("migrations", io.session.Migrations()),
					zap.Error(err))
			}
		}
	}

	io.barrier.Close()
	io.barrier.Wait()

	io.dataChan.Close()
	io.eventChan.Close()

	for buff := range io.dataChan.Out() {
		if err := io.session.trans.SendData(buff.Payload()); err != nil {
			log.L(io.session.gate.svcCtx).Error("session send data error",
				zap.String("session_id", io.session.Id().String()),
				zap.String("local", io.session.NetAddr().Local.String()),
				zap.String("remote", io.session.NetAddr().Remote.String()),
				zap.Int64("migrations", io.session.Migrations()),
				zap.Error(err))
		}
		buff.Release()
	}

	for event := range io.eventChan.Out() {
		err := transport.Retry{
			Transceiver: &io.session.transceiver,
			Times:       io.session.gate.options.IORetryTimes,
		}.Send(io.session.transceiver.Send(event))
		if err != nil {
			log.L(io.session.gate.svcCtx).Error("session send event failed",
				zap.String("session_id", io.session.Id().String()),
				zap.String("local", io.session.NetAddr().Local.String()),
				zap.String("remote", io.session.NetAddr().Remote.String()),
				zap.Int64("migrations", io.session.Migrations()),
				zap.Error(err))
		}
	}

	async.ReturnVoid(io.terminated)
}

func (io *_SessionIO) handlePayload(event transport.Event[*gtp.MsgPayload]) {
	rejected := io.dataListeners.Broadcast(event.Msg.Data)
	if rejected > 0 {
		log.L(io.session.gate.svcCtx).Error("some listeners rejected the receive payload due to backpressure",
			zap.String("session_id", io.session.Id().String()),
			zap.Uint32("seq", event.Seq),
			zap.Uint32("ack", event.Ack),
			zap.Int("rejected", rejected))
	}
}

func (io *_SessionIO) handleEvent(event transport.IEvent) {
	rejected := io.eventListeners.Broadcast(event)
	if rejected > 0 {
		log.L(io.session.gate.svcCtx).Error("some listeners rejected the receive event due to backpressure",
			zap.String("session_id", io.session.Id().String()),
			zap.Uint32("seq", event.Seq),
			zap.Uint32("ack", event.Ack),
			zap.Uint8("msg_id", event.Msg.MsgId()),
			zap.Int("rejected", rejected))
	}
}

type _SessionDataIO _SessionIO

// Send 发送数据
func (io *_SessionDataIO) Send(data []byte) error {
	if !io.barrier.Join(1) {
		return errors.New("gate: session data i/o is terminating")
	}
	defer io.barrier.Done()

	io.dataChan.In() <- binaryutil.CloneBytes(true, data)
	return nil
}

// Listen 监听数据
func (io *_SessionDataIO) Listen(ctx context.Context, handler SessionDataHandler) error {
	if handler == nil {
		return errors.New("gate: handler is nil")
	}
	return io.addListener(ctx, handler)
}

func (io *_SessionDataIO) addListener(ctx context.Context, handler SessionDataHandler) error {
	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case <-io.session.Done():
		return errors.New("gate: session data i/o is terminating")
	default:
	}

	if !io.barrier.Join(1) {
		return errors.New("gate: session data i/o is terminating")
	}

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-ctx.Done():
		case <-io.session.Done():
		}
		cancel()
	}()

	listener := io.dataListeners.Add(handler, io.session.gate.options.SessionDataListenerInboxSize)

	go func() {
		defer io.barrier.Done()
		for {
			select {
			case <-ctx.Done():
				io.dataListeners.Delete(listener)
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

type _SessionEventIO _SessionIO

// Send 发送事件
func (io *_SessionEventIO) Send(event transport.IEvent) error {
	if !io.barrier.Join(1) {
		return errors.New("gate: session event i/o is terminating")
	}
	defer io.barrier.Done()

	io.eventChan.In() <- event
	return nil
}

// Listen 监听事件
func (io *_SessionEventIO) Listen(ctx context.Context, handler SessionEventHandler) error {
	if handler == nil {
		return errors.New("gate: handler is nil")
	}
	return io.addListener(ctx, handler)
}

func (io *_SessionEventIO) addListener(ctx context.Context, handler SessionEventHandler) error {
	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case <-io.session.Done():
		return errors.New("gate: session event i/o is terminating")
	default:
	}

	if !io.barrier.Join(1) {
		return errors.New("gate: session event i/o is terminating")
	}

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-ctx.Done():
		case <-io.session.Done():
		}
		cancel()
	}()

	listener := io.eventListeners.Add(handler, io.session.gate.options.SessionEventListenerInboxSize)

	go func() {
		defer io.barrier.Done()
		for {
			select {
			case <-ctx.Done():
				io.eventListeners.Delete(listener)
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
