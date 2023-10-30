package nats_broker

import (
	"context"
	"fmt"
	"github.com/nats-io/nats.go"
	"kit.golaxy.org/plugins/broker"
	"kit.golaxy.org/plugins/internal"
	"kit.golaxy.org/plugins/log"
	"strings"
)

func (b *_Broker) newSubscriber(ctx context.Context, mode _SubscribeMode, pattern string, opts broker.SubscriberOptions) (*_Subscriber, error) {
	if b.options.TopicPrefix != "" {
		pattern = b.options.TopicPrefix + pattern
	}

	ctx, cancel := context.WithCancel(ctx)

	sub := &_Subscriber{
		broker:              b,
		ctx:                 ctx,
		cancel:              cancel,
		stoppedChan:         make(chan struct{}, 1),
		unsubscribedHandler: opts.UnsubscribedHandler,
	}

	var err error

	if opts.QueueName != "" {
		queueName := opts.QueueName
		if b.options.QueuePrefix != "" {
			queueName = b.options.QueuePrefix + queueName
		}
		sub.natsSub, err = b.client.QueueSubscribe(pattern, queueName, sub.handleMsg)
	} else {
		sub.natsSub, err = b.client.Subscribe(pattern, sub.handleMsg)
	}

	if err != nil {
		return nil, fmt.Errorf("%w: %w", broker.ErrBroker, err)
	}

	switch mode {
	case _SubscribeMode_Sync, _SubscribeMode_Chan:
		sub.eventChan = make(chan broker.Event, opts.EventChanSize)
	case _SubscribeMode_Handler:
		sub.eventHandler = opts.EventHandler
	}

	go sub.run()

	log.Infof(b.ctx, "subscribe topic pattern %q with queue %q", pattern, sub.natsSub.Queue)

	return sub, nil
}

type _SubscribeMode int32

const (
	_SubscribeMode_Handler _SubscribeMode = iota
	_SubscribeMode_Sync
	_SubscribeMode_Chan
)

type _Subscriber struct {
	broker              *_Broker
	ctx                 context.Context
	cancel              context.CancelFunc
	stoppedChan         chan struct{}
	natsSub             *nats.Subscription
	eventChan           chan broker.Event
	eventHandler        broker.EventHandler
	unsubscribedHandler broker.UnsubscribedHandler
}

// Pattern returns the subscription pattern used to create the subscriber.
func (s *_Subscriber) Pattern() string {
	return strings.TrimPrefix(s.natsSub.Subject, s.broker.options.TopicPrefix)
}

// QueueName subscribers with the same queue name will create a shared subscription where each receives a subset of messages.
func (s *_Subscriber) QueueName() string {
	return strings.TrimPrefix(s.natsSub.Queue, s.broker.options.QueuePrefix)
}

// Unsubscribe unsubscribes the subscriber from the topic.
func (s *_Subscriber) Unsubscribe() <-chan struct{} {
	s.cancel()
	return s.stoppedChan
}

// Next is a blocking call that waits for the next event to be received from the subscriber.
func (s *_Subscriber) Next() (broker.Event, error) {
	for event := range s.eventChan {
		return event, nil
	}
	return nil, broker.ErrUnsubscribed
}

// EventChan returns a channel that can be used to receive events from the subscriber.
func (s *_Subscriber) EventChan() <-chan broker.Event {
	return s.eventChan
}

func (s *_Subscriber) run() {
	<-s.ctx.Done()
	defer func() { s.stoppedChan <- struct{}{} }()

	if err := s.natsSub.Unsubscribe(); err != nil {
		log.Errorf(s.broker.ctx, "unsubscribe topic pattern %q with %q failed, %s", s.natsSub.Subject, s.natsSub.Queue, err)
	} else {
		log.Infof(s.broker.ctx, "unsubscribe topic pattern %q with %q success", s.natsSub.Subject, s.natsSub.Queue)
	}

	if s.eventChan != nil {
		close(s.eventChan)
	}

	if s.unsubscribedHandler != nil {
		s.unsubscribedHandler(s)
	}
}

func (s *_Subscriber) handleMsg(msg *nats.Msg) {
	e := &_NatsEvent{
		msg: msg,
		ns:  s,
	}

	if s.eventHandler != nil {
		if err := internal.Call(func() error { return s.eventHandler(e) }); err != nil {
			log.Errorf(s.broker.ctx, "handle msg event failed, %s, nak: %s", err, e.Nak(context.Background()))
		}
	}

	if s.eventChan != nil {
		select {
		case s.eventChan <- e:
		default:
			log.Errorf(s.broker.ctx, "handle msg event failed, event chan is full, nak: %s", e.Nak(context.Background()))
		}
	}
}
