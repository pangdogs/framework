package transport

import (
	"errors"
	"kit.golaxy.org/golaxy/util/generic"
	"kit.golaxy.org/plugins/gtp"
)

type (
	PayloadHandler = generic.DelegateFunc1[Event[gtp.MsgPayload], error] // Payload消息事件处理器
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
	return t.retrySend(t.Transceiver.Send(Event[gtp.MsgPayload]{
		Msg: gtp.MsgPayload{Data: data},
	}.Pack()))
}

func (t *TransProtocol) retrySend(err error) error {
	return Retry{
		Transceiver: t.Transceiver,
		Times:       t.RetryTimes,
	}.Send(err)
}

// HandleEvent 消息事件处理器
func (t *TransProtocol) HandleEvent(e Event[gtp.Msg]) error {
	switch e.Msg.MsgId() {
	case gtp.MsgId_Payload:
		return t.PayloadHandler.Exec(func(err, _ error) bool {
			return err == nil || !errors.Is(err, ErrUnexpectedMsg)
		}, UnpackEvent[gtp.MsgPayload](e))
	default:
		return ErrUnexpectedMsg
	}
}
