package protocol

import (
	"errors"
	"kit.golaxy.org/plugins/transport"
)

type (
	HandlePayload = func(Event[*transport.MsgPayload]) error
)

// TransProtocol 传输协议
type TransProtocol struct {
	Transceiver   *Transceiver  // 消息事件收发器
	HandlePayload HandlePayload // Payload消息事件句柄
}

// SendPayload 发送数据
func (t *TransProtocol) SendPayload(data []byte, sequenced bool) error {
	if t.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	return t.Transceiver.Send(PackEvent(Event[*transport.MsgPayload]{
		Flags: transport.Flags_None().Setd(transport.Flag_Sequenced, sequenced),
		Msg:   &transport.MsgPayload{Data: data},
	}))
}

// HandleEvent 消息事件处理句柄
func (t *TransProtocol) HandleEvent(e Event[transport.Msg]) error {
	switch e.Msg.MsgId() {
	case transport.MsgId_Payload:
		payload := UnpackEvent[*transport.MsgPayload](e)

		if t.HandlePayload != nil {
			return t.HandlePayload(payload)
		}

		return nil
	default:
		return ErrUnexpectedMsg
	}
}
