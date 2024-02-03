package transport

import (
	"errors"
	"git.golaxy.org/core/util/generic"
	"git.golaxy.org/framework/net/gtp"
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
		var errs []error

		t.PayloadHandler.Exec(func(err, _ error) bool {
			if err != nil {
				errs = append(errs, err)
			}
			return false
		}, UnpackEvent[gtp.MsgPayload](e))

		if len(errs) > 0 {
			return errors.Join(errs...)
		}

		return nil

	default:
		return nil
	}
}
