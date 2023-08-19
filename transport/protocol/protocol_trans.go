package protocol

import (
	"errors"
	"kit.golaxy.org/plugins/transport"
)

type (
	PayloadHandler = func(Event[*transport.MsgPayload]) error // Payload消息事件处理器
)

// TransProtocol 传输协议
type TransProtocol struct {
	Transceiver    *Transceiver   // 消息事件收发器
	RetryTimes     int            // 网络io超时时的重试次数
	PayloadHandler PayloadHandler // Payload消息事件处理器
}

// SendData 发送数据
func (t *TransProtocol) SendData(data []byte) error {
	if t.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	return t.retrySend(t.Transceiver.Send(PackEvent(Event[*transport.MsgPayload]{
		Msg: &transport.MsgPayload{Data: data},
	})))
}

func (t *TransProtocol) retrySend(err error) error {
	return Retry{
		Transceiver: t.Transceiver,
		Times:       t.RetryTimes,
	}.Send(err)
}

// EventHandler 消息事件处理器
func (t *TransProtocol) EventHandler(e Event[transport.Msg]) error {
	switch e.Msg.MsgId() {
	case transport.MsgId_Payload:
		payload := UnpackEvent[*transport.MsgPayload](e)

		if t.PayloadHandler != nil {
			return t.PayloadHandler(payload)
		}

		return nil
	default:
		return ErrUnexpectedMsg
	}
}
