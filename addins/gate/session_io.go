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
	"git.golaxy.org/framework/utils/fanout"
	"go.uber.org/zap"
)

type (
	// SessionDataHandler 处理会话收到的一段原始负载。
	SessionDataHandler = generic.DelegateVoid2[ISession, []byte]
	// SessionEventHandler 处理会话收到的一个 GTP 事件。
	SessionEventHandler = generic.DelegateVoid2[ISession, transport.IEvent]
)

// IDataIO 提供会话原始负载的异步发送与监听。
type IDataIO interface {
	// Send 复制 data 并将其加入发送队列。
	Send(data []byte) error
	// Listen 注册监听器，直到 ctx 取消或会话关闭。
	Listen(ctx context.Context, handler SessionDataHandler) error
}

// IEventIO 提供会话 GTP 事件的异步发送与监听。
type IEventIO interface {
	// Send 将 event 加入发送队列；调用方须保证事件在处理完成前保持有效。
	Send(event transport.IEvent) error
	// Listen 注册监听器，直到 ctx 取消或会话关闭。
	Listen(ctx context.Context, handler SessionEventHandler) error
}

type _SessionIO struct {
	session        *_Session
	barrier        generic.Barrier
	terminated     async.Completer
	dataChan       *generic.UnboundedChannel[binaryutil.Bytes]
	eventChan      *generic.UnboundedChannel[transport.IEvent]
	dataListeners  fanout.Broadcaster[SessionDataHandler, []byte]
	eventListeners fanout.Broadcaster[SessionEventHandler, transport.IEvent]
}

func (io *_SessionIO) init(session *_Session) {
	io.session = session
	io.terminated, _ = async.NewSignal()
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
					zap.String("session_id", io.session.ID().String()),
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
					zap.String("session_id", io.session.ID().String()),
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
				zap.String("session_id", io.session.ID().String()),
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
				zap.String("session_id", io.session.ID().String()),
				zap.String("local", io.session.NetAddr().Local.String()),
				zap.String("remote", io.session.NetAddr().Remote.String()),
				zap.Int64("migrations", io.session.Migrations()),
				zap.Error(err))
		}
	}

	io.terminated.Complete()
}

func (io *_SessionIO) handlePayload(event transport.Event[*gtp.MsgPayload]) {
	dropped := io.dataListeners.Broadcast(event.Msg.Data)
	if dropped > 0 {
		log.L(io.session.gate.svcCtx).Error("received payload deliveries dropped due to listener backpressure",
			zap.String("session_id", io.session.ID().String()),
			zap.Uint32("seq", event.Seq),
			zap.Uint32("ack", event.Ack),
			zap.Int("dropped", dropped))
	}
}

func (io *_SessionIO) handleEvent(event transport.IEvent) {
	dropped := io.eventListeners.Broadcast(event)
	if dropped > 0 {
		log.L(io.session.gate.svcCtx).Error("received event deliveries dropped due to listener backpressure",
			zap.String("session_id", io.session.ID().String()),
			zap.Uint32("seq", event.Seq),
			zap.Uint32("ack", event.Ack),
			zap.Uint8("msg_id", event.Msg.MsgID()),
			zap.Int("dropped", dropped))
	}
}

type _SessionDataIO _SessionIO

// Send 复制 data 后加入发送队列；成功仅表示已入队，会话 I/O 正在停止时返回错误。
func (io *_SessionDataIO) Send(data []byte) error {
	if !io.barrier.Join(1) {
		return errors.New("gate: session data i/o is terminating")
	}
	defer io.barrier.Done()

	io.dataChan.In() <- binaryutil.CloneBytes(true, data)
	return nil
}

// Listen 注册原始负载处理器，直到 ctx 取消或会话关闭；handler 为 nil 时返回错误。
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

	listener := io.dataListeners.Subscribe(handler, io.session.gate.options.SessionDataListenerInboxSize)

	go func() {
		defer io.barrier.Done()
		for {
			select {
			case <-ctx.Done():
				io.dataListeners.Unsubscribe(listener)
				log.L(io.session.gate.svcCtx).Debug("delete a session data listener", zap.String("session_id", io.session.ID().String()))
				return
			case data := <-listener.Inbox:
				listener.Handler.Call(io.session.gate.svcCtx.AutoRecover(), io.session.gate.svcCtx.ReportError(), func(panicError error) bool {
					if panicError != nil {
						log.L(io.session.gate.svcCtx).Error("handle session data panicked",
							zap.String("session_id", io.session.ID().String()),
							zap.Error(panicError))
					}
					return false
				}, io.session, data)
			}
		}
	}()

	log.L(io.session.gate.svcCtx).Debug("add a session data listener", zap.String("session_id", io.session.ID().String()))
	return nil
}

type _SessionEventIO _SessionIO

// Send 将 event 加入发送队列；成功仅表示已入队，调用方须在处理完成前保持事件有效。
func (io *_SessionEventIO) Send(event transport.IEvent) error {
	if !io.barrier.Join(1) {
		return errors.New("gate: session event i/o is terminating")
	}
	defer io.barrier.Done()

	io.eventChan.In() <- event
	return nil
}

// Listen 注册 GTP 事件处理器，直到 ctx 取消或会话关闭；handler 为 nil 时返回错误。
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

	listener := io.eventListeners.Subscribe(handler, io.session.gate.options.SessionEventListenerInboxSize)

	go func() {
		defer io.barrier.Done()
		for {
			select {
			case <-ctx.Done():
				io.eventListeners.Unsubscribe(listener)
				log.L(io.session.gate.svcCtx).Debug("delete a session event listener", zap.String("session_id", io.session.ID().String()))
				return
			case event := <-listener.Inbox:
				listener.Handler.Call(io.session.gate.svcCtx.AutoRecover(), io.session.gate.svcCtx.ReportError(), func(panicError error) bool {
					if panicError != nil {
						log.L(io.session.gate.svcCtx).Error("handle session event panicked",
							zap.String("session_id", io.session.ID().String()),
							zap.Error(panicError))
					}
					return false
				}, io.session, event)
			}
		}
	}()

	log.L(io.session.gate.svcCtx).Debug("add a session event listener", zap.String("session_id", io.session.ID().String()))
	return nil
}
