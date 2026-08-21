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

package dsvc

import (
	"context"
	"errors"

	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/addins/broker"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/net/gap"
	"go.uber.org/zap"
)

type (
	// MsgHandler 处理一条已解码的分布式服务消息；第一个参数是接收话题。
	MsgHandler = generic.DelegateVoid2[string, gap.MsgPacket]
)

type _BrokerMsg struct {
	topic     string
	queue     string
	msgPacket gap.MsgPacket
}

func (d *_DistService) addListener(ctx context.Context, handler MsgHandler) (async.Signal, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case <-d.ctx.Done():
		return async.Signal{}, errors.New("dsvc: dsvc is terminating")
	default:
	}

	if !d.barrier.Join(1) {
		return async.Signal{}, errors.New("dsvc: dsvc is terminating")
	}

	listener := d.listeners.Subscribe(handler, d.options.ListenerInboxSize)
	stopped, stoppedSignal := async.NewSignal()

	go func() {
		defer d.barrier.Done()
		defer log.L(d.svcCtx).Debug("delete a broker message listener")
		defer stopped.Complete()
		defer d.listeners.Unsubscribe(listener)

		for {
			select {
			case <-ctx.Done():
				return
			case <-d.ctx.Done():
				return
			case msg := <-listener.Inbox:
				listener.Handler.Call(d.svcCtx.AutoRecover(), d.svcCtx.ReportError(), func(panicError error) bool {
					if panicError != nil {
						log.L(d.svcCtx).Error("handle decoded broker message panicked",
							zap.String("topic", msg.topic),
							zap.String("queue", msg.queue),
							zap.Error(panicError))
					}
					return false
				}, msg.topic, msg.msgPacket)
			}
		}
	}()

	log.L(d.svcCtx).Debug("add a broker message listener")
	return stoppedSignal, nil
}

func (d *_DistService) handleEvent(e broker.Event) {
	mp, err := d.decoder.Decode(e.Message)
	if err != nil {
		log.L(d.svcCtx).Error("decode broker message failed",
			zap.String("topic", e.Topic),
			zap.String("queue", e.Queue),
			zap.Error(err))
		return
	}

	msg := _BrokerMsg{
		topic:     e.Topic,
		queue:     e.Queue,
		msgPacket: mp,
	}

	dropped := d.listeners.Broadcast(msg)
	if dropped > 0 {
		log.L(d.svcCtx).Error("broker message deliveries dropped due to listener backpressure",
			zap.String("topic", e.Topic),
			zap.String("queue", e.Queue),
			zap.Int("dropped", dropped))
	}
}
