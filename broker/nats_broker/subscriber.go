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

type _SubscribeMode int32

const (
	_SubscribeMode_Handler _SubscribeMode = iota
	_SubscribeMode_Sync
	_SubscribeMode_Chan
)

func (b *_Broker) newSubscriber(ctx context.Context, mode _SubscribeMode, pattern string, opts broker.SubscriberOptions) (*_Subscriber, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if b.options.TopicPrefix != "" {
		pattern = b.options.TopicPrefix + pattern
	}

	ctx, cancel := context.WithCancel(ctx)

	sub := &_Subscriber{
		Context:             ctx,
		broker:              b,
		cancel:              cancel,
		stoppedChan:         make(chan struct{}),
		unsubscribedHandler: opts.UnsubscribedHandler,
	}

	var handleMsg nats.MsgHandler

	switch mode {
	case _SubscribeMode_Sync, _SubscribeMode_Chan:
		sub.eventChan = make(chan broker.Event, opts.EventChanSize)
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
		return nil, fmt.Errorf("%w: %w", broker.ErrBroker, err)
	}

	log.Debugf(b.ctx, "subscribe topic pattern %q queue %q success", sub.Queue(), sub.Queue())

	b.wg.Add(1)
	go sub.mainLoop()

	return sub, nil
}

type _Subscriber struct {
	context.Context
	broker              *_Broker
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
func (s *_Subscriber) EventChan() (<-chan broker.Event, error) {
	return s.eventChan, nil
}

func (s *_Subscriber) mainLoop() {
	defer func() {
		s.cancel()
		s.broker.wg.Done()
		close(s.stoppedChan)
	}()

	select {
	case <-s.Done():
	case <-s.broker.ctx.Done():
	}

	if err := s.natsSub.Unsubscribe(); err != nil {
		log.Errorf(s.broker.ctx, "unsubscribe topic pattern %q with %q failed, %s", s.Pattern(), s.Queue(), err)
	} else {
		log.Debugf(s.broker.ctx, "unsubscribe topic pattern %q with %q success", s.Pattern(), s.Queue())
	}

	if s.eventChan != nil {
		close(s.eventChan)
	}

	s.unsubscribedHandler.Invoke(func(panicErr error) bool {
		log.Errorf(s.broker.ctx, "handle unsubscribed topic pattern %q queue %q failed, %s", s.Pattern(), s.Queue(), panicErr)
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
		log.Errorf(s.broker.ctx, "handle msg from topic %q queue %q failed, event chan is full, nak: %v", e.Topic(), e.Queue(), nakErr)
	}
}

func (s *_Subscriber) handleEventProcess(msg *nats.Msg) {
	e := &_Event{
		msg: msg,
		ns:  s,
	}

	s.eventHandler.Invoke(func(err error, panicErr error) bool {
		if err := generic.FuncError(err, panicErr); err != nil {
			log.Errorf(s.broker.ctx, "handle msg from topic %q queue %q failed, %s", e.Topic(), e.Queue(), err)
		}
		return panicErr != nil
	}, e)
}
