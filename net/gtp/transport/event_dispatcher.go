package transport

import (
	"context"
	"errors"
	"git.golaxy.org/core/utils/generic"
)

type (
	EventHandler = generic.DelegateFunc1[IEvent, error] // 消息事件处理器
	ErrorHandler = generic.DelegateAction1[error]       // 错误处理器
)

// EventDispatcher 消息事件分发器
type EventDispatcher struct {
	Transceiver  *Transceiver // 消息事件收发器
	RetryTimes   int          // 网络io超时时的重试次数
	EventHandler EventHandler // 消息事件处理器列表
}

// Dispatching 分发事件
func (d *EventDispatcher) Dispatching(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if d.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}

	defer d.Transceiver.GC()

	e, err := d.retryRecv(ctx)
	if err != nil {
		return err
	}

	var errs []error

	d.EventHandler.Invoke(func(err, panicErr error) bool {
		if err := generic.FuncError(err, panicErr); err != nil {
			errs = append(errs, err)
		}
		return panicErr != nil
	}, e)

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// Run 运行
func (d *EventDispatcher) Run(ctx context.Context, errorHandler ...ErrorHandler) {
	if ctx == nil {
		ctx = context.Background()
	}

	var _errorHandler ErrorHandler
	if len(errorHandler) > 0 {
		_errorHandler = errorHandler[0]
	}

	if d.Transceiver == nil {
		_errorHandler.Invoke(nil, errors.New("setting Transceiver is nil"))
		return
	}

	defer d.Transceiver.Clean()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := d.Dispatching(ctx); err != nil {
			_errorHandler.Invoke(nil, err)
		}
	}
}

func (d *EventDispatcher) retryRecv(ctx context.Context) (IEvent, error) {
	e, err := d.Transceiver.Recv(ctx)
	return Retry{
		Transceiver: d.Transceiver,
		Times:       d.RetryTimes,
		Ctx:         ctx,
	}.Recv(e, err)
}
