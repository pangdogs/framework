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
	"fmt"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/addins/broker"
	"git.golaxy.org/framework/addins/log"
	"github.com/nats-io/nats.go"
	"strings"
)

type _SubscribeMode int32

const (
	_SubscribeMode_Handler _SubscribeMode = iota
	_SubscribeMode_Sync
	_SubscribeMode_Chan
)

type _SubscriberSettings struct {
	broker  *_Broker
	ctx     context.Context
	pattern string
}

// With applies additional settings to the subscriber.
func (s *_SubscriberSettings) With(settings ...option.Setting[broker.SubscriberOptions]) (broker.ISubscriber, error) {
	return s.broker.Subscribe(s.ctx, s.pattern, settings...)
}

type _SyncSubscriberSettings struct {
	broker  *_Broker
	ctx     context.Context
	pattern string
}

// With applies additional settings to the subscriber.
func (s *_SyncSubscriberSettings) With(settings ...option.Setting[broker.SubscriberOptions]) (broker.ISyncSubscriber, error) {
	return s.broker.SubscribeSync(s.ctx, s.pattern, settings...)
}

type _ChanSubscriberSettings struct {
	broker  *_Broker
	ctx     context.Context
	pattern string
}

// With applies additional settings to the subscriber.
func (s *_ChanSubscriberSettings) With(settings ...option.Setting[broker.SubscriberOptions]) (broker.IChanSubscriber, error) {
	return s.broker.SubscribeChan(s.ctx, s.pattern, settings...)
}

func (b *_Broker) newSubscriber(ctx context.Context, mode _SubscribeMode, pattern string, opts broker.SubscriberOptions) (*_Subscriber, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if b.options.TopicPrefix != "" {
		pattern = b.options.TopicPrefix + pattern
	}

	ctx, cancel := context.WithCancel(ctx)

	sub := &_Subscriber{
		Context:        ctx,
		terminate:      cancel,
		terminated:     async.MakeAsyncRet(),
		broker:         b,
		unsubscribedCB: opts.UnsubscribedCB,
	}

	var handleMsg nats.MsgHandler

	switch mode {
	case _SubscribeMode_Sync, _SubscribeMode_Chan:
		sub.eventChan = make(chan broker.IEvent, opts.EventChanSize)
		handleMsg = sub.handleEventChan
	case _SubscribeMode_Handler:
		sub.eventHandler = opts.EventHandler
		handleMsg = sub.handleEventProcess
	}

	var err error

	if opts.Queue != "" {
		queue := opts.Queue
		if b.options.QueuePrefix != "" {
			queue = b.options.QueuePrefix + queue
		}
		sub.natsSub, err = b.client.QueueSubscribe(pattern, queue, handleMsg)
	} else {
		sub.natsSub, err = b.client.Subscribe(pattern, handleMsg)
	}

	if err != nil {
		return nil, fmt.Errorf("broker: %w", err)
	}

	log.Debugf(b.svcCtx, "subscribe topic pattern %q queue %q success", pattern, sub.Queue())

	b.wg.Add(1)
	go sub.mainLoop()

	return sub, nil
}

type _Subscriber struct {
	context.Context
	terminate      context.CancelFunc
	terminated     chan async.Ret
	broker         *_Broker
	natsSub        *nats.Subscription
	eventChan      chan broker.IEvent
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

// Next is a blocking call that waits for the next event to be received from the subscriber.
func (s *_Subscriber) Next() (broker.IEvent, error) {
	for event := range s.eventChan {
		return event, nil
	}
	return nil, broker.ErrUnsubscribed
}

// EventChan returns a channel that can be used to receive events from the subscriber.
func (s *_Subscriber) EventChan() (<-chan broker.IEvent, error) {
	return s.eventChan, nil
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
		log.Errorf(s.broker.svcCtx, "unsubscribe topic pattern %q with %q failed, %s", s.Pattern(), s.Queue(), err)
	} else {
		log.Debugf(s.broker.svcCtx, "unsubscribe topic pattern %q with %q success", s.Pattern(), s.Queue())
	}

	if s.eventChan != nil {
		close(s.eventChan)
	}

	s.unsubscribedCB.SafeCall(func(panicErr error) bool {
		log.Errorf(s.broker.svcCtx, "handle unsubscribed topic pattern %q queue %q failed, %s", s.Pattern(), s.Queue(), panicErr)
		return false
	}, s)
}

func (s *_Subscriber) handleEventChan(msg *nats.Msg) {
	e := &_Event{
		msg: msg,
		ns:  s,
	}

	select {
	case s.eventChan <- e:
	default:
		var nakErr error
		if e.Queue() != "" {
			nakErr = e.Nak(context.Background())
		}
		log.Errorf(s.broker.svcCtx, "handle msg from topic %q queue %q failed, receive event chan is full, nak: %v", e.Topic(), e.Queue(), nakErr)
	}
}

func (s *_Subscriber) handleEventProcess(msg *nats.Msg) {
	e := &_Event{
		msg: msg,
		ns:  s,
	}

	s.eventHandler.SafeCall(func(err error, panicErr error) bool {
		if err := generic.FuncError(err, panicErr); err != nil {
			log.Errorf(s.broker.svcCtx, "handle msg from topic %q queue %q failed, %s", e.Topic(), e.Queue(), err)
		}
		return panicErr != nil
	}, e)
}
