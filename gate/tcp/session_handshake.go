package tcp

import (
	"kit.golaxy.org/plugins/gate"
)

type _TcpSessionHandshake struct {
	*_TcpSession
}

// StateChangedHandler 设置接收会话状态变化的处理器
func (s *_TcpSessionHandshake) StateChangedHandler(handler gate.StateChangedHandler) error {
	s.stateChangedHandler = handler
	return nil
}

// RecvHandler 设置接收数据的处理器
func (s *_TcpSessionHandshake) RecvHandler(handler gate.RecvHandler) error {
	s.recvHandler = handler
	return nil
}

// RecvChanSize 设置接收数据的chan的大小，<=0表示不使用chan
func (s *_TcpSessionHandshake) RecvChanSize(size int) error {
	if size <= 0 {
		s.recvChan = nil
		return nil
	}

	s.recvChan = make(chan gate.Recv, size)

	return nil
}

// RecvEventHandlers 设置接收自定义事件的处理器
func (s *_TcpSessionHandshake) RecvEventHandlers(handlers []gate.RecvEventHandler) error {
	s.dispatcher.EventHandlers = append(s.dispatcher.EventHandlers, handlers...)
	return nil
}

// RecvEventSize 设置自定义事件的chan的大小，<=0表示不使用chan
func (s *_TcpSessionHandshake) RecvEventSize(size int) error {
	if size <= 0 {
		s.recvEventChan = nil
		return nil
	}

	s.recvEventChan = make(chan gate.RecvEvent, size)

	return nil
}
