package broker

import (
	"fmt"
	"golang.org/x/net/context"
	"kit.golaxy.org/golaxy"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util/generic"
	"kit.golaxy.org/plugins/log"
)

// MakeWriteChan creates a new channel for publishing data to a specific topic.
func MakeWriteChan(servCtx service.Context, topic string, size int) chan<- []byte {
	ch := make(chan []byte, size)

	go func() {
		defer func() {
			if info := recover(); info != nil {
				log.Errorf(servCtx, "%s: publish data to topic %q failed, %s", golaxy.ErrPanicked, topic, info)
			}
		}()

		broker := Using(servCtx)

		for {
			select {
			case data, ok := <-ch:
				if !ok {
					return
				}
				if err := broker.Publish(servCtx, topic, data); err != nil {
					log.Errorf(servCtx, "publish data to topic %q failed, %s", topic, err)
				}
			case <-servCtx.Done():
				return
			}
		}
	}()

	return ch
}

// MakeReadChan creates a new channel for receiving data from a specific pattern.
func MakeReadChan(servCtx service.Context, ctx context.Context, pattern, queue string, size int) (<-chan []byte, error) {
	ch := make(chan []byte, size)

	_, err := Using(servCtx).Subscribe(ctx, pattern,
		Option{}.Queue(queue),
		Option{}.EventHandler(generic.CastDelegateFunc1(func(e IEvent) error {
			select {
			case ch <- e.Message():
				return nil
			default:
				var nakErr error
				if e.Queue() != "" {
					nakErr = e.Nak(context.Background())
				}
				return fmt.Errorf("read chan is full, nak: %v", nakErr)
			}
		})),
		Option{}.UnsubscribedHandler(generic.CastDelegateAction1(func(sub ISubscriber) {
			close(ch)
		})))
	if err != nil {
		return nil, err
	}

	return ch, nil
}
