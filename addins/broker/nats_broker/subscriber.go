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

package nats_broker

import (
	"context"
	"errors"
	"fmt"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/addins/broker"
	"git.golaxy.org/framework/addins/log"
	"github.com/nats-io/nats.go"
	"strings"
)

func (b *_Broker) newSubscriber(ctx context.Context, pattern string, opts broker.SubscriberOptions) (*_Subscriber, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if b.options.TopicPrefix != "" {
		pattern = b.options.TopicPrefix + pattern
	}

	ctx, cancel := context.WithCancel(ctx)

	s := &_Subscriber{
		Context:        ctx,
		terminate:      cancel,
		terminated:     async.MakeAsyncRet(),
		broker:         b,
		eventHandler:   opts.EventHandler,
		unsubscribedCB: opts.UnsubscribedCB,
	}

	if opts.EventChanSize > 0 {
		s.eventChan = make(chan broker.Event, opts.EventChanSize)
	}

	var err error

	if opts.Queue != "" {
		queue := opts.Queue
		if b.options.QueuePrefix != "" {
			queue = b.options.QueuePrefix + queue
		}
		s.natsSub, err = b.client.QueueSubscribe(pattern, queue, s.handleMsg)
	} else {
		s.natsSub, err = b.client.Subscribe(pattern, s.handleMsg)
	}

	if err != nil {
		return nil, fmt.Errorf("broker: %w", err)
	}

	log.Debugf(b.svcCtx, "subscribe topic pattern %q with queue %q success", pattern, s.Queue())

	b.wg.Add(1)
	go s.mainLoop()

	return s, nil
}

type _Subscriber struct {
	context.Context
	terminate      context.CancelFunc
	terminated     chan async.Ret
	broker         *_Broker
	natsSub        *nats.Subscription
	eventChan      chan broker.Event
	eventHandler   broker.EventHandler
	unsubscribedCB broker.UnsubscribedCB
}

// Pattern returns the subscription pattern used to create the subscriber.
func (s *_Subscriber) Pattern() string {
	return strings.TrimPrefix(s.natsSub.Subject, s.broker.options.TopicPrefix)
}

// Queue subscribers with the same queue name will create a shared subscription where each receives a subset of messages.
func (s *_Subscriber) Queue() string {
	return strings.TrimPrefix(s.natsSub.Queue, s.broker.options.QueuePrefix)
}

// Unsubscribe unsubscribes the subscriber from the topic.
func (s *_Subscriber) Unsubscribe() async.AsyncRet {
	s.terminate()
	return s.terminated
}

// Unsubscribed subscriber is unsubscribed.
func (s *_Subscriber) Unsubscribed() async.AsyncRet {
	return s.terminated
}

// EventChan returns a channel that can be used to receive events from the subscriber.
func (s *_Subscriber) EventChan() <-chan broker.Event {
	if s.eventChan == nil {
		log.Panicf(s.broker.svcCtx, "event channel size less equal 0, can't be used")
	}
	return s.eventChan
}

func (s *_Subscriber) mainLoop() {
	defer func() {
		s.terminate()
		s.broker.wg.Done()
		async.Return(s.terminated, async.VoidRet)
	}()

	select {
	case <-s.Done():
	case <-s.broker.ctx.Done():
	}

	if err := s.natsSub.Unsubscribe(); err != nil {
		log.Errorf(s.broker.svcCtx, "unsubscribe topic pattern %q with queue %q failed, %s", s.Pattern(), s.Queue(), err)
	} else {
		log.Debugf(s.broker.svcCtx, "unsubscribe topic pattern %q with queue %q success", s.Pattern(), s.Queue())
	}

	if s.eventChan != nil {
		close(s.eventChan)
	}

	s.unsubscribedCB.SafeCall(s)
}

func (s *_Subscriber) handleMsg(msg *nats.Msg) {
	e := broker.Event{
		Pattern: s.Pattern(),
		Topic:   strings.TrimPrefix(msg.Subject, s.broker.options.TopicPrefix),
		Queue:   s.Queue(),
		Message: msg.Data,
		Ack:     unsupportedAck,
		Nak:     unsupportedNak,
	}

	if s.eventChan != nil {
		select {
		case s.eventChan <- e:
		default:
			log.Errorf(s.broker.svcCtx, "handle msg from topic %q with queue %q failed, receive event chan is full", e.Topic, e.Queue)
		}
	}

	if s.eventHandler != nil {
		s.eventHandler.SafeCall(func(err error, panicErr error) bool {
			if err := generic.FuncError(err, panicErr); err != nil {
				log.Errorf(s.broker.svcCtx, "handle msg from topic %q with queue %q failed, %s", e.Topic, e.Queue, err)
			}
			return panicErr != nil
		}, e)
	}
}

func unsupportedAck(ctx context.Context) error {
	return errors.New("broker: used not JetStream, unable to acknowledge(ack)")
}

func unsupportedNak(ctx context.Context) error {
	return errors.New("broker: used not JetStream, unable to negatively acknowledge(nak)")
}
