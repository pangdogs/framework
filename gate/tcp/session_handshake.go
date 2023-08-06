package tcp

import (
	"kit.golaxy.org/plugins/gate"
)

// _TcpSessionHandshake 会话握手设置
type _TcpSessionHandshake struct {
	*_TcpSession
}

// InitStateChangedHandlers 设置接收会话状态变化的处理器
func (s *_TcpSessionHandshake) InitStateChangedHandlers(handlers []gate.StateChangedHandler) error {
	s.stateChangedHandlers = handlers
	return nil
}

// InitRecvDataHandlers 设置接收数据的处理器
func (s *_TcpSessionHandshake) InitRecvDataHandlers(handlers []gate.RecvDataHandler) error {
	s.recvDataHandlers = handlers
	return nil
}

// InitRecvEventHandlers 设置接收自定义事件的处理器
func (s *_TcpSessionHandshake) InitRecvEventHandlers(handlers []gate.RecvEventHandler) error {
	s.recvEventhandlers = handlers
	return nil
}

// InitRecvDataChanSize 设置接收数据的chan的大小，<=0表示不使用chan
func (s *_TcpSessionHandshake) InitRecvDataChanSize(size int) error {
	if size <= 0 {
		s.recvDataChan = nil
		return nil
	}
	s.recvDataChan = make(chan gate.RecvData, size)
	return nil
}

// InitRecvEventSize 设置自定义事件的chan的大小，<=0表示不使用chan
func (s *_TcpSessionHandshake) InitRecvEventSize(size int) error {
	if size <= 0 {
		s.recvEventChan = nil
		return nil
	}
	s.recvEventChan = make(chan gate.RecvEvent, size)
	return nil
}
