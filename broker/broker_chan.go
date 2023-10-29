package broker

import (
	"golang.org/x/net/context"
	"kit.golaxy.org/golaxy"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/plugins/log"
)

// MakeInputChan creates a new channel for publishing data to a specific topic.
func MakeInputChan(serviceCtx service.Context, topic string, size int) chan<- []byte {
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

// MakeOutputChan creates a new channel for receiving data from a specific pattern.
func MakeOutputChan(serviceCtx service.Context, ctx context.Context, pattern, queue string, size int) (<-chan []byte, error) {
	ch := make(chan []byte, size)

	_, err := Using(serviceCtx).Subscribe(ctx, pattern,
		Option{}.QueueName(queue),
		Option{}.EventHandler(func(e Event) error {
			select {
			case ch <- e.Message():
			default:
				log.Errorf(serviceCtx, "receive data from topic %q failed, output chan is full, nak: %s", e.Topic(), pattern, e.Nak(context.Background()))
			}
			return nil
		}),
		Option{}.UnsubscribedHandler(func(sub Subscriber) {
			close(ch)
		}))
	if err != nil {
		return nil, err
	}

	return ch, nil
}
