package gtp

import (
	"kit.golaxy.org/plugins/gate"
)

// _TcpSessionSetting 会话设置
type _TcpSessionSetting struct {
	*_TcpSession
}

// InitStateChangedHandlers 设置接收会话状态变化的处理器
func (s *_TcpSessionSetting) InitStateChangedHandlers(handlers []gate.StateChangedHandler) error {
	s.stateChangedHandlers = handlers
	return nil
}

// InitRecvDataHandlers 设置接收数据的处理器
func (s *_TcpSessionSetting) InitRecvDataHandlers(handlers []gate.RecvDataHandler) error {
	s.recvDataHandlers = handlers
	return nil
}

// InitRecvEventHandlers 设置接收自定义事件的处理器
func (s *_TcpSessionSetting) InitRecvEventHandlers(handlers []gate.RecvEventHandler) error {
	s.recvEventHandlers = handlers
	return nil
}

// InitRecvDataChanSize 设置接收数据的chan的大小，<=0表示不使用chan
func (s *_TcpSessionSetting) InitRecvDataChanSize(size int) error {
	if size <= 0 {
		s.recvDataChan = nil
		return nil
	}
	s.recvDataChan = make(chan gate.RecvData, size)
	return nil
}

// InitRecvEventSize 设置自定义事件的chan的大小，<=0表示不使用chan
func (s *_TcpSessionSetting) InitRecvEventSize(size int) error {
	if size <= 0 {
		s.recvEventChan = nil
		return nil
	}
	s.recvEventChan = make(chan gate.RecvEvent, size)
	return nil
}
