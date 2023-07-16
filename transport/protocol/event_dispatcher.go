package protocol

import (
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"kit.golaxy.org/plugins/transport"
)

var (
	ErrHandlerNotRegistered = errors.New("handler not registered") // 消息处理句柄未注册
	ErrUnexpectedMsg        = errors.New("unexpected msg")         // 收到非预期的消息
)

// EventHandler 消息事件句柄
type EventHandler interface {
	HandleEvent(Event[transport.Msg]) error
}

// ErrorHandler 错误处理句柄
type ErrorHandler = func(err error)

// EventDispatcher 消息事件分发器
type EventDispatcher struct {
	Transceiver   *Transceiver   // 消息事件收发器
	EventHandlers []EventHandler // 消息处理句柄
	ErrorHandler  ErrorHandler   // 错误处理句柄
}

// Bind 绑定句柄
func (d *EventDispatcher) Bind(handler EventHandler) error {
	if handler == nil {
		return errors.New("handler is nil")
	}
	d.EventHandlers = append(d.EventHandlers, handler)
	return nil
}

// Unbind 解绑定句柄
func (d *EventDispatcher) Unbind(handler EventHandler) error {
	if handler == nil {
		return errors.New("handler is nil")
	}
	for i := range d.EventHandlers {
		if d.EventHandlers[i] == handler {
			d.EventHandlers = append(d.EventHandlers[:i], d.EventHandlers[i+1:]...)
			break
		}
	}
	return nil
}

// Run 运行
func (d *EventDispatcher) Run(ctx context.Context) {
	if d.Transceiver == nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		e, err := d.Transceiver.Recv()
		if err != nil {
			if d.ErrorHandler != nil {
				d.ErrorHandler(err)
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
					d.ErrorHandler(err)
				}
			}
			handled = true
			break
		}

		if !handled {
			if d.ErrorHandler != nil {
				d.ErrorHandler(fmt.Errorf("%w: %d", ErrHandlerNotRegistered, e.Msg.MsgId()))
			}
		}
	}
}
