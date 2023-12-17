package transport

import (
	"context"
	"errors"
	"fmt"
	"kit.golaxy.org/golaxy"
	"kit.golaxy.org/golaxy/util/generic"
	"kit.golaxy.org/plugins/gtp"
)

var (
	ErrUnableToDispatch = errors.New("gtp: unable to dispatch") // 无法分发消息事件
	ErrUnexpectedMsg    = errors.New("gtp: unexpected msg")     // 收到非预期的消息
)

type (
	EventHandler = generic.DelegateFunc1[Event[gtp.Msg], error] // 消息事件处理器
	ErrorHandler = generic.DelegateAction1[error]               // 错误处理器
)

// EventDispatcher 消息事件分发器
type EventDispatcher struct {
	Transceiver  *Transceiver // 消息事件收发器
	RetryTimes   int          // 网络io超时时的重试次数
	EventHandler EventHandler // 消息事件处理器列表
}

// Dispatching 分发事件
func (d *EventDispatcher) Dispatching() error {
	if d.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}

	defer d.Transceiver.GC()

	e, err := d.retryRecv(d.Transceiver.Recv())
	if err != nil {
		return err
	}

	err = generic.FuncError(d.EventHandler.Invoke(func(err, panicErr error) bool {
		err = generic.FuncError(err, panicErr)
		if err == nil || !errors.Is(err, ErrUnexpectedMsg) {
			return true
		}
		return false
	}, e))
	if err == nil {
		return nil
	}

	return fmt.Errorf("%w (%d)", ErrUnableToDispatch, e.Msg.MsgId())
}

// Run 运行
func (d *EventDispatcher) Run(ctx context.Context, errorHandler ErrorHandler) {
	if d.Transceiver == nil {
		errorHandler.Invoke(nil, errors.New("setting Transceiver is nil"))
		return
	}

	if ctx == nil {
		errorHandler.Invoke(nil, fmt.Errorf("%w: ctx is nil", golaxy.ErrArgs))
		return
	}

	defer d.Transceiver.Clean()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := d.Dispatching(); err != nil {
			errorHandler.Invoke(nil, err)
		}
	}
}

func (d *EventDispatcher) retryRecv(e Event[gtp.Msg], err error) (Event[gtp.Msg], error) {
	return Retry{
		Transceiver: d.Transceiver,
		Times:       d.RetryTimes,
	}.Recv(e, err)
}
