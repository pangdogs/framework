package protocol

import (
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"kit.golaxy.org/plugins/transport"
)

var (
	ErrHandlerNotRegistered = errors.New("handler not registered") // 消息处理器未注册
	ErrUnexpectedMsg        = errors.New("unexpected msg")         // 收到非预期的消息
)

// EventHandler 消息事件处理器
type EventHandler interface {
	// HandleEvent 消息事件处理句柄
	HandleEvent(Event[transport.Msg]) error
}

// ErrorHandler 错误处理器
type ErrorHandler = func(ctx context.Context, err error)

// EventDispatcher 消息事件分发器
type EventDispatcher struct {
	Transceiver   *Transceiver   // 消息事件收发器
	RetryTimes    int            // 网络io超时时的重试次数
	EventHandlers []EventHandler // 消息事件处理器
	ErrorHandler  ErrorHandler   // 错误处理器
}

// Add 添加消息事件处理器
func (d *EventDispatcher) Add(handler EventHandler) error {
	if handler == nil {
		return errors.New("handler is nil")
	}

	for i := range d.EventHandlers {
		if d.EventHandlers[i] == handler {
			return nil
		}
	}

	d.EventHandlers = append(d.EventHandlers, handler)

	return nil
}

// Remove 删除消息事件处理器
func (d *EventDispatcher) Remove(handler EventHandler) error {
	if handler == nil {
		return errors.New("handler is nil")
	}

	for i := range d.EventHandlers {
		if d.EventHandlers[i] == handler {
			d.EventHandlers = append(d.EventHandlers[:i], d.EventHandlers[i+1:]...)
			return nil
		}
	}

	return errors.New("handler not found")
}

// Run 运行
func (d *EventDispatcher) Run(ctx context.Context) {
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

		e, err := d.retryRecv(d.Transceiver.Recv())
		if err != nil {
			if d.ErrorHandler != nil {
				d.ErrorHandler(ctx, err)
			}
			continue
		}

		handled := false

		for i := range d.EventHandlers {
			if err = d.EventHandlers[i].HandleEvent(e); err != nil {
				if errors.Is(err, ErrUnexpectedMsg) {
					continue
				}
				if d.ErrorHandler != nil {
					d.ErrorHandler(ctx, err)
				}
			}
			handled = true
			break
		}

		if !handled {
			if d.ErrorHandler != nil {
				d.ErrorHandler(ctx, fmt.Errorf("%w: %d", ErrHandlerNotRegistered, e.Msg.MsgId()))
			}
		}

		d.Transceiver.GC()
	}
}

func (d *EventDispatcher) retryRecv(e Event[transport.Msg], err error) (Event[transport.Msg], error) {
	return Retry{
		Transceiver: d.Transceiver,
		Times:       d.RetryTimes,
	}.Recv(e, err)
}
