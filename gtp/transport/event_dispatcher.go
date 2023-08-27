package transport

import (
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"kit.golaxy.org/plugins/gtp"
	"kit.golaxy.org/plugins/internal"
)

var (
	ErrHandlerNotRegistered = errors.New("handler not registered") // 消息处理器未注册
	ErrUnexpectedMsg        = errors.New("unexpected msg")         // 收到非预期的消息
)

type (
	EventHandler = func(Event[gtp.Msg]) error // 消息事件处理器
	ErrorHandler = func(err error)            // 错误处理器
)

// EventDispatcher 消息事件分发器
type EventDispatcher struct {
	Transceiver   *Transceiver   // 消息事件收发器
	RetryTimes    int            // 网络io超时时的重试次数
	EventHandlers []EventHandler // 消息事件处理器
}

// Dispatching 分发事件
func (d *EventDispatcher) Dispatching() error {
	if d.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}

	d.Transceiver.GC()

	e, err := d.retryRecv(d.Transceiver.Recv())
	if err != nil {
		return err
	}

	for i := range d.EventHandlers {
		if err = internal.Call(func() error { return d.EventHandlers[i](e) }); err != nil {
			if errors.Is(err, ErrUnexpectedMsg) {
				continue
			}
			return err
		}
		return nil
	}

	return fmt.Errorf("%w: %d", ErrHandlerNotRegistered, e.Msg.MsgId())
}

// Run 运行
func (d *EventDispatcher) Run(ctx context.Context, errorHandler ErrorHandler) {
	if d.Transceiver == nil {
		return
	}

	defer d.Transceiver.GC()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := d.Dispatching(); err != nil {
			if errorHandler != nil {
				internal.CallVoid(func() { errorHandler(err) })
			}
		}
	}
}

func (d *EventDispatcher) retryRecv(e Event[gtp.Msg], err error) (Event[gtp.Msg], error) {
	return Retry{
		Transceiver: d.Transceiver,
		Times:       d.RetryTimes,
	}.Recv(e, err)
}
