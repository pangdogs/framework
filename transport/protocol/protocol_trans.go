package protocol

import (
	"errors"
	"fmt"
	"kit.golaxy.org/plugins/transport"
)

type (
	RecvPayload = func(Event[*transport.MsgPayload]) error
)

// TransProtocol 传输协议
type TransProtocol struct {
	Transceiver *Transceiver // 消息事件收发器
	RecvPayload RecvPayload  // 接收Payload消息事件
}

// SendPayload 发送Payload消息事件
func (t *TransProtocol) SendPayload(e Event[*transport.MsgPayload]) error {
	if t.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	return t.Transceiver.Send(PackEvent(e))
}

// Bind 绑定事件分发器
func (t *TransProtocol) Bind(dispatcher *Dispatcher) error {
	if dispatcher == nil {
		return errors.New("dispatcher is nil")
	}

	if dispatcher.Handlers == nil {
		dispatcher.Handlers = map[transport.MsgId]Handler{}
	}

	dispatcher.Handlers[transport.MsgId_Payload] = t

	return nil
}

// Unbind 解绑定事件分发器
func (t *TransProtocol) Unbind(dispatcher *Dispatcher) error {
	if dispatcher == nil {
		return errors.New("dispatcher is nil")
	}

	if dispatcher.Handlers == nil {
		return nil
	}

	if dispatcher.Handlers[transport.MsgId_Payload] == t {
		delete(dispatcher.Handlers, transport.MsgId_Payload)
	}

	return nil
}

// Recv 消息事件处理句柄
func (t *TransProtocol) Recv(e Event[transport.Msg]) error {
	switch e.Msg.MsgId() {
	case transport.MsgId_Payload:
		payload := UnpackEvent[*transport.MsgPayload](e)

		if t.RecvPayload != nil {
			return t.RecvPayload(payload)
		}

		return nil
	default:
		return fmt.Errorf("%w: %d", ErrRecvUnexpectedMsg, e.Msg.MsgId())
	}
}
