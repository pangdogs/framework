package gtp

import (
	"errors"
	"kit.golaxy.org/plugins/gate"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/protocol"
)

// GetSessionSetting 获取会话设置接口
func GetSessionSetting(session gate.Session) (gate.SessionSetting, error) {
	gtpSession, ok := session.(*_GtpSession)
	if !ok {
		return nil, errors.New("incorrect session type")
	}

	switch session.GetState() {
	case gate.SessionState_Handshake, gate.SessionState_Confirmed:
		return &_GtpSessionSetting{_GtpSession: gtpSession}, nil
	default:
		return nil, errors.New("incorrect session state")
	}
}

// _GtpSessionSetting 会话设置
type _GtpSessionSetting struct {
	*_GtpSession
}

// StateChangedHandlers 设置接收会话状态变化的处理器
func (s *_GtpSessionSetting) StateChangedHandlers(handlers ...gate.StateChangedHandler) error {
	s.stateChangedHandlers = handlers
	return nil
}

// RecvDataHandlers 设置接收数据的处理器
func (s *_GtpSessionSetting) RecvDataHandlers(handlers ...gate.RecvDataHandler) error {
	s.recvDataHandlers = handlers
	return nil
}

// RecvEventHandlers 设置接收自定义事件的处理器
func (s *_GtpSessionSetting) RecvEventHandlers(handlers ...gate.RecvEventHandler) error {
	s.recvEventHandlers = handlers
	return nil
}

// SendDataChanSize 设置发送数据的channel的大小，<=0表示不使用channel
func (s *_GtpSessionSetting) SendDataChanSize(size int) error {
	if size <= 0 {
		s.sendDataChan = nil
		return nil
	}
	s.sendDataChan = make(chan gate.SendData, size)
	return nil
}

// RecvDataChanSize 设置接收数据的channel的大小，<=0表示不使用channel
func (s *_GtpSessionSetting) RecvDataChanSize(size int) error {
	if size <= 0 {
		s.recvDataChan = nil
		return nil
	}
	s.recvDataChan = make(chan gate.RecvData, size)
	return nil
}

// SendEventSize 设置发送自定义事件的channel的大小，<=0表示不使用channel
func (s *_GtpSessionSetting) SendEventSize(size int) error {
	if size <= 0 {
		s.sendEventChan = nil
		return nil
	}
	s.sendEventChan = make(chan protocol.Event[transport.Msg], size)
	return nil
}

// RecvEventSize 设置自定义事件的channel的大小，<=0表示不使用channel
func (s *_GtpSessionSetting) RecvEventSize(size int) error {
	if size <= 0 {
		s.recvEventChan = nil
		return nil
	}
	s.recvEventChan = make(chan gate.RecvEvent, size)
	return nil
}
