package broker

import (
	"golang.org/x/net/context"
	"kit.golaxy.org/golaxy"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util/generic"
	"kit.golaxy.org/plugins/log"
)

// MakeWriteChan creates a new channel for publishing data to a specific topic.
func MakeWriteChan(serviceCtx service.Context, topic string, size int) chan<- []byte {
	ch := make(chan []byte, size)

	go func() {
		defer func() {
			if info := recover(); info != nil {
				log.Errorf(serviceCtx, "%s: publish data to topic %q failed, %s", golaxy.ErrPanicked, topic, info)
			}
		}()

		broker := Using(serviceCtx)

		for {
			select {
			case data, ok := <-ch:
				if !ok {
					return
				}
				if err := broker.Publish(serviceCtx, topic, data); err != nil {
					log.Errorf(serviceCtx, "publish data to topic %q failed, %s", topic, err)
				}
			case <-serviceCtx.Done():
				return
			}
		}
	}()

	return ch
}

// MakeReadChan creates a new channel for receiving data from a specific pattern.
func MakeReadChan(serviceCtx service.Context, ctx context.Context, pattern, queue string, size int) (<-chan []byte, error) {
	ch := make(chan []byte, size)

	_, err := Using(serviceCtx).Subscribe(ctx, pattern,
		Option{}.Queue(queue),
		Option{}.EventHandler(generic.CastDelegateFunc1(func(e Event) error {
			select {
			case ch <- e.Message():
			default:
				var nakErr error
				if e.Queue() != "" {
					nakErr = e.Nak(context.Background())
				}
				log.Errorf(serviceCtx, "receive data from topic %q queue %q failed, output chan is full, nak: %s", e.Topic(), e.Queue(), nakErr)
			}
			return nil
		})),
		Option{}.UnsubscribedHandler(generic.CastDelegateAction1(func(sub Subscriber) {
			close(ch)
		})))
	if err != nil {
		return nil, err
	}

	return ch, nil
}
