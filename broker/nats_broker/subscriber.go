package nats_broker

import (
	"context"
	"fmt"
	"github.com/nats-io/nats.go"
	"kit.golaxy.org/golaxy/util/generic"
	"kit.golaxy.org/plugins/broker"
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

	if opts.Queue != "" {
		queue := opts.Queue
		if b.options.QueuePrefix != "" {
			queue = b.options.QueuePrefix + queue
		}
		sub.natsSub, err = b.client.QueueSubscribe(pattern, queue, sub.handleMsg)
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

	go sub.mainLoop()

	log.Debugf(b.ctx, "subscribe topic pattern %q queue %q success", sub.Queue(), sub.Queue())

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

// Queue subscribers with the same queue name will create a shared subscription where each receives a subset of messages.
func (s *_Subscriber) Queue() string {
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

func (s *_Subscriber) mainLoop() {
	<-s.ctx.Done()

	defer func() { s.stoppedChan <- struct{}{} }()

	if err := s.natsSub.Unsubscribe(); err != nil {
		log.Errorf(s.broker.ctx, "unsubscribe topic pattern %q with %q failed, %s", s.Pattern(), s.Queue(), err)
	} else {
		log.Debugf(s.broker.ctx, "unsubscribe topic pattern %q with %q success", s.Pattern(), s.Queue())
	}

	if s.eventChan != nil {
		close(s.eventChan)
	}

	if err := s.unsubscribedHandler.Invoke(nil, s); err != nil {
		log.Errorf(s.broker.ctx, "handle unsubscribed from topic pattern %q queue %q failed, %s", s.Pattern(), s.Queue(), err)
	}
}

func (s *_Subscriber) handleMsg(msg *nats.Msg) {
	e := &_Event{
		msg: msg,
		ns:  s,
	}

	if err := generic.FuncError(s.eventHandler.Invoke(nil, e)); err != nil {
		var nakErr error
		if e.Queue() != "" {
			nakErr = e.Nak(context.Background())
		}
		log.Errorf(s.broker.ctx, "handle msg from topic %q queue %q failed, %s, nak: %s", e.Topic(), e.Queue(), err, nakErr)
	}

	if s.eventChan != nil {
		select {
		case s.eventChan <- e:
		default:
			var nakErr error
			if e.Queue() != "" {
				nakErr = e.Nak(context.Background())
			}
			log.Errorf(s.broker.ctx, "handle msg from topic %q queue %q failed, event chan is full, nak: %s", e.Topic(), e.Queue(), nakErr)
		}
	}
}
