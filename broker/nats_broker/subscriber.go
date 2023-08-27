package nats_broker

import (
	"github.com/nats-io/nats.go"
	"golang.org/x/net/context"
	"kit.golaxy.org/plugins/broker"
	"kit.golaxy.org/plugins/internal"
	"kit.golaxy.org/plugins/logger"
	"strings"
)

type _SubscribeMode int32

const (
	_SubscribeMode_Handler _SubscribeMode = iota
	_SubscribeMode_Sync
	_SubscribeMode_Chan
)

func newNatsSubscriber(ctx context.Context, nb *_NatsBroker, mode _SubscribeMode, pattern string, opts broker.SubscriberOptions) (*_NatsSubscriber, error) {
	if nb.options.TopicPrefix != "" {
		pattern = nb.options.TopicPrefix + pattern
	}

	queueName := opts.QueueName
	if queueName != "" && nb.options.QueuePrefix != "" {
		queueName = nb.options.QueuePrefix + queueName
	}

	var sub *nats.Subscription
	var err error
	var eventChan chan broker.Event
	var eventHandler broker.EventHandler

	switch mode {
	case _SubscribeMode_Handler:
		eventHandler = opts.EventHandler
	default:
		eventChan = make(chan broker.Event, opts.EventChanSize)
	}

	ns := &_NatsSubscriber{}

	msgHandler := func(msg *nats.Msg) {
		e := &_NatsEvent{
			msg: msg,
			ns:  ns,
		}

		switch mode {
		case _SubscribeMode_Handler:
			if eventHandler != nil {
				if err := internal.Call(func() error { return eventHandler(e) }); err != nil {
					logger.Tracef(ns.nb.ctx, "handle msg event failed, %s", err)
				}
			}
		default:
			select {
			case eventChan <- e:
			default:
				logger.Trace(ns.nb.ctx, "msg event chan is full")
			}
		}
	}

	if queueName != "" {
		sub, err = nb.client.QueueSubscribe(pattern, queueName, msgHandler)
	} else {
		sub, err = nb.client.Subscribe(pattern, msgHandler)
	}
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	ns.cancel = cancel
	ns.nb = nb
	ns.sub = sub
	ns.options = opts
	ns.eventChan = eventChan

	go func() {
		<-ctx.Done()
		if err := sub.Unsubscribe(); err != nil {
			logger.Errorf(nb.ctx, "unsubscribe topic %q with %q failed, %s", sub.Subject, sub.Queue, err)
		} else {
			logger.Debugf(nb.ctx, "unsubscribe topic %q with %q", sub.Subject, sub.Queue)
		}
		if eventChan != nil {
			close(eventChan)
		}
		if opts.UnsubscribedCb != nil {
			opts.UnsubscribedCb(ns)
		}
	}()

	logger.Debugf(nb.ctx, "subscribe topic %q with queue %q", pattern, queueName)

	return ns, nil
}

type _NatsSubscriber struct {
	cancel    context.CancelFunc
	nb        *_NatsBroker
	sub       *nats.Subscription
	options   broker.SubscriberOptions
	eventChan chan broker.Event
}

// Pattern returns the subscription pattern used to create the subscriber.
func (s *_NatsSubscriber) Pattern() string {
	return strings.TrimPrefix(s.sub.Subject, s.nb.options.TopicPrefix)
}

// QueueName subscribers with the same queue name will create a shared subscription where each receives a subset of messages.
func (s *_NatsSubscriber) QueueName() string {
	return strings.TrimPrefix(s.sub.Queue, s.nb.options.QueuePrefix)
}

// Unsubscribe unsubscribes the subscriber from the topic.
func (s *_NatsSubscriber) Unsubscribe() {
	s.cancel()
}

// Next is a blocking call that waits for the next event to be received from the subscriber.
func (s *_NatsSubscriber) Next() (broker.Event, error) {
	for event := range s.eventChan {
		return event, nil
	}
	return nil, broker.ErrUnsubscribed
}

// EventChan returns a channel that can be used to receive events from the subscriber.
func (s *_NatsSubscriber) EventChan() <-chan broker.Event {
	return s.eventChan
}
