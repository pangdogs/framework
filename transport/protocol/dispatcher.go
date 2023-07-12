package protocol

import (
	"errors"
	"golang.org/x/net/context"
	"kit.golaxy.org/plugins/transport"
)

var (
	ErrHandlerNotRegistered = errors.New("handler not registered")   // 消息处理句柄未注册
	ErrRecvUnexpectedMsg    = errors.New("recv unexpected msg")      // 收到非预期的消息
	ErrRecvUnexpectedSeq    = errors.New("recv unexpected sequence") // 收到非预期的消息序号
)

// ErrorHandler 错误句柄
type ErrorHandler = func(err error) bool

// Handler 消息句柄
type Handler interface {
	Recv(Event[transport.Msg]) error
}

// Dispatcher 消息事件分发器
type Dispatcher struct {
	Transceiver *Transceiver                // 消息事件收发器
	Handlers    map[transport.MsgId]Handler // 消息处理句柄
}

// Run 运行
func (d *Dispatcher) Run(ctx context.Context, errorHandler ErrorHandler) {
	if d.Transceiver == nil {
		if errorHandler != nil {
			errorHandler(errors.New("setting Transceiver is nil"))
		}
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
			if errorHandler != nil && !errorHandler(err) {
				return
			}
			continue
		}

		if e.Flags.Is(transport.Flag_Sequenced) {
			if e.Seq < d.Transceiver.RecvSeq {
				continue
			} else if e.Seq > d.Transceiver.RecvSeq {
				if errorHandler != nil && !errorHandler(ErrRecvUnexpectedSeq) {
					return
				}
				continue
			}
		}

		var handler Handler

		if d.Handlers != nil {
			handler = d.Handlers[e.Msg.MsgId()]
		}

		if handler == nil {
			if errorHandler != nil && !errorHandler(ErrHandlerNotRegistered) {
				return
			}
			continue
		}

		err = handler.Recv(e)
		if err != nil {
			if errorHandler != nil && !errorHandler(err) {
				return
			}
			continue
		}
	}
}
