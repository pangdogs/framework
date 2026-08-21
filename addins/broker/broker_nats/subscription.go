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

package broker_nats

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/addins/broker"
	"git.golaxy.org/framework/addins/log"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// addSubscriber 创建事件流或回调订阅，并用 barrier 防止关闭期间出现半注册资源。
func (b *_NatsBroker) addSubscriber(ctx context.Context, pattern, queue string, handler broker.EventHandler) (<-chan broker.Event, async.Signal, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case <-b.scope.Context().Done():
		return nil, async.Signal{}, errors.New("broker: broker is terminating")
	default:
	}

	if !b.barrier.Join(1) {
		return nil, async.Signal{}, errors.New("broker: broker is terminating")
	}
	defer b.barrier.Done()

	natsPattern := pattern
	if b.options.TopicPrefix != "" {
		natsPattern = b.options.TopicPrefix + natsPattern
	}

	natsQueue := queue
	if natsQueue != "" {
		if b.options.QueuePrefix != "" {
			natsQueue = b.options.QueuePrefix + natsQueue
		}
	}

	var eventChan *generic.UnboundedChannel[broker.Event]
	if handler == nil {
		eventChan = generic.NewUnboundedChannel[broker.Event]()
	}

	handleMsg := func(msg *nats.Msg) {
		event := broker.Event{
			Pattern: pattern,
			Topic:   strings.TrimPrefix(msg.Subject, b.options.TopicPrefix),
			Queue:   queue,
			Message: msg.Data,
			Ack:     unsupportedAck,
			Nak:     unsupportedNak,
		}

		if eventChan != nil {
			eventChan.In() <- event
		}

		if handler != nil {
			handler.Call(b.svcCtx.AutoRecover(), b.svcCtx.ReportError(), func(panicErr error) bool {
				if panicErr != nil {
					log.L(b.svcCtx).Error("handle msg from topic panicked",
						zap.String("topic", msg.Subject),
						zap.String("pattern", natsPattern),
						zap.String("queue", natsQueue),
						zap.Error(panicErr))
				}
				return false
			}, event)
		}
	}

	var natsSub *nats.Subscription
	var err error

	if queue != "" {
		natsSub, err = b.client.QueueSubscribe(pattern, queue, handleMsg)
	} else {
		natsSub, err = b.client.Subscribe(pattern, handleMsg)
	}

	if err != nil {
		if eventChan != nil {
			eventChan.Close()
		}

		log.L(b.svcCtx).Error("subscribe topic pattern failed", zap.String("pattern", natsPattern), zap.String("queue", natsQueue), zap.Error(err))
		return nil, async.Signal{}, fmt.Errorf("broker: %w", err)
	}

	unsubscribed, unsubscribedSignal := async.NewSignal()
	cleanup := func() {
		defer unsubscribed.Complete()
		if eventChan != nil {
			defer eventChan.Close()
		}
		if unsubscribeErr := natsSub.Unsubscribe(); unsubscribeErr != nil {
			log.L(b.svcCtx).Error("unsubscribe topic pattern failed", zap.String("pattern", natsPattern), zap.String("queue", natsQueue), zap.Error(unsubscribeErr))
		} else {
			log.L(b.svcCtx).Debug("unsubscribe topic pattern ok", zap.String("pattern", natsPattern), zap.String("queue", natsQueue))
		}
	}

	future := async.SpawnVoid(b.scope, func(scopeCtx context.Context) {
		select {
		case <-ctx.Done():
		case <-scopeCtx.Done():
		}
		cleanup()
	})
	future.OnComplete(func(ret async.Result) {
		if ret.Error != nil && !errors.Is(ret.Error, async.ErrScopeClosed) {
			log.L(b.svcCtx).Error("broker subscription task failed", zap.Error(ret.Error))
		}
	})

	if ret, ok := future.TryGet(); ok && errors.Is(ret.Error, async.ErrScopeClosed) {
		cleanup()
		return nil, async.Signal{}, errors.New("broker: broker is terminating")
	}

	log.L(b.svcCtx).Debug("subscribe topic pattern ok", zap.String("pattern", natsPattern), zap.String("queue", natsQueue))

	if eventChan != nil {
		return eventChan.Out(), unsubscribedSignal, nil
	}
	return nil, unsubscribedSignal, nil
}

func unsupportedAck(ctx context.Context) error {
	return errors.New("broker: used not JetStream, unable to acknowledge(ack)")
}

func unsupportedNak(ctx context.Context) error {
	return errors.New("broker: used not JetStream, unable to negatively acknowledge(nak)")
}
