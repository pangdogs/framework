package broker

import (
	"context"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/utils/binaryutil"
)

type (
	ErrorHandler = generic.DelegateAction1[error] // 错误处理器
)

// MakeWriteChan creates a new channel for publishing data to a specific topic.
func MakeWriteChan(broker IBroker, topic string, size int, errorHandler ...ErrorHandler) chan<- binaryutil.RecycleBytes {
	if broker == nil {
		panic(fmt.Errorf("%w: broker is nil", core.ErrArgs))
	}

	var _errorHandler ErrorHandler
	if len(errorHandler) > 0 {
		_errorHandler = errorHandler[0]
	}

	ch := make(chan binaryutil.RecycleBytes, size)

	go func() {
		defer func() {
			for bs := range ch {
				bs.Release()
			}
		}()
		for bs := range ch {
			err := broker.Publish(context.Background(), topic, bs.Data())
			bs.Release()
			if err != nil {
				_errorHandler.Invoke(nil, err)
			}
		}
	}()

	return ch
}

// MakeReadChan creates a new channel for receiving data from a specific pattern.
func MakeReadChan(broker IBroker, ctx context.Context, pattern, queue string, size int) (<-chan binaryutil.RecycleBytes, error) {
	if broker == nil {
		panic(fmt.Errorf("%w: broker is nil", core.ErrArgs))
	}

	if ctx == nil {
		ctx = context.Background()
	}

	ch := make(chan binaryutil.RecycleBytes, size)

	_, err := broker.Subscribe(ctx, pattern,
		With.Queue(queue),
		With.EventHandler(generic.MakeDelegateFunc1(func(e IEvent) error {
			bs := binaryutil.MakeNonRecycleBytes(e.Message())

			select {
			case ch <- bs:
				return nil
			default:
				var nakErr error
				if e.Queue() != "" {
					nakErr = e.Nak(context.Background())
				}
				return fmt.Errorf("read chan is full, nak: %v", nakErr)
			}
		})),
		With.UnsubscribedCB(generic.MakeDelegateAction1(func(sub ISubscriber) {
			close(ch)
		})))
	if err != nil {
		return nil, err
	}

	return ch, nil
}
