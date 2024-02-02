package broker

import (
	"context"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/util/generic"
	"strings"
)

type (
	ErrorHandler = generic.DelegateAction1[error] // 错误处理器
)

// Path return topic path.
func Path(broker IBroker, elems ...string) string {
	if broker == nil {
		panic(fmt.Errorf("%w: broker is nil", core.ErrArgs))
	}
	return strings.Join(elems, broker.GetSeparator())
}

// MakeWriteChan creates a new channel for publishing data to a specific topic.
func MakeWriteChan(broker IBroker, ctx context.Context, topic string, size int, errorHandler ...ErrorHandler) chan<- []byte {
	if broker == nil {
		panic(fmt.Errorf("%w: broker is nil", core.ErrArgs))
	}

	if ctx == nil {
		ctx = context.Background()
	}

	var _errorHandler ErrorHandler
	if len(errorHandler) > 0 {
		_errorHandler = errorHandler[0]
	}

	ch := make(chan []byte, size)

	go func() {
		for {
			select {
			case data, ok := <-ch:
				if !ok {
					return
				}
				if err := broker.Publish(ctx, topic, data); err != nil {
					_errorHandler.Invoke(nil, err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch
}

// MakeReadChan creates a new channel for receiving data from a specific pattern.
func MakeReadChan(broker IBroker, ctx context.Context, pattern, queue string, size int) (<-chan []byte, error) {
	if broker == nil {
		panic(fmt.Errorf("%w: broker is nil", core.ErrArgs))
	}

	if ctx == nil {
		ctx = context.Background()
	}

	ch := make(chan []byte, size)

	_, err := broker.Subscribe(ctx, pattern,
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
