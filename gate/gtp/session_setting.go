package gtp

import (
	"errors"
	"kit.golaxy.org/plugins/gate"
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

// InitStateChangedHandlers 设置接收会话状态变化的处理器
func (s *_GtpSessionSetting) InitStateChangedHandlers(handlers []gate.StateChangedHandler) error {
	s.stateChangedHandlers = handlers
	return nil
}

// InitRecvDataHandlers 设置接收数据的处理器
func (s *_GtpSessionSetting) InitRecvDataHandlers(handlers []gate.RecvDataHandler) error {
	s.recvDataHandlers = handlers
	return nil
}

// InitRecvEventHandlers 设置接收自定义事件的处理器
func (s *_GtpSessionSetting) InitRecvEventHandlers(handlers []gate.RecvEventHandler) error {
	s.recvEventHandlers = handlers
	return nil
}

// InitRecvDataChanSize 设置接收数据的chan的大小，<=0表示不使用chan
func (s *_GtpSessionSetting) InitRecvDataChanSize(size int) error {
	if size <= 0 {
		s.recvDataChan = nil
		return nil
	}
	s.recvDataChan = make(chan gate.RecvData, size)
	return nil
}

// InitRecvEventSize 设置自定义事件的chan的大小，<=0表示不使用chan
func (s *_GtpSessionSetting) InitRecvEventSize(size int) error {
	if size <= 0 {
		s.recvEventChan = nil
		return nil
	}
	s.recvEventChan = make(chan gate.RecvEvent, size)
	return nil
}
