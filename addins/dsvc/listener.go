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

	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/addins/broker"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/net/gap"
	"go.uber.org/zap"
)

type (
	MsgHandler = generic.DelegateVoid2[string, gap.MsgPacket] // 消息处理器
)

type _BrokerMsg struct {
	topic     string
	queue     string
	msgPacket gap.MsgPacket
}

func (d *_DistService) addListener(ctx context.Context, handler MsgHandler) error {
	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case <-d.ctx.Done():
		return errors.New("dsvc: dsvc is terminating")
	default:
	}

	if !d.barrier.Join(1) {
		return errors.New("dsvc: dsvc is terminating")
	}

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-ctx.Done():
		case <-d.ctx.Done():
		}
		cancel()
	}()

	listener := d.listeners.Add(handler, d.options.ListenerInboxSize)

	go func() {
		defer d.barrier.Done()

		for {
			select {
			case <-ctx.Done():
				d.listeners.Delete(listener)
				log.L(d.svcCtx).Debug("delete a broker message listener")
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
	return nil
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

	rejected := d.listeners.Broadcast(msg)
	if rejected > 0 {
		log.L(d.svcCtx).Error("some listeners rejected the broker message due to backpressure",
			zap.String("topic", e.Topic),
			zap.String("queue", e.Queue),
			zap.Int("rejected", rejected))
	}
}
