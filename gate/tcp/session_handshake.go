package tcp

import (
	"errors"
	"kit.golaxy.org/plugins/gate"
)

// StateChangedHandler 设置接收会话状态变化的处理器
func (s *_TcpSession) StateChangedHandler(handler gate.StateChangedHandler) error {
	if s.state != gate.SessionState_Handshake {
		return errors.New("session state is not handshake")
	}

	s.stateChangedHandler = handler

	return nil
}

// RecvHandler 设置接收数据的处理器
func (s *_TcpSession) RecvHandler(handler gate.RecvHandler) error {
	if s.state != gate.SessionState_Handshake {
		return errors.New("session state is not handshake")
	}

	s.recvHandler = handler

	return nil
}

// RecvChanSize 设置接收数据的chan的大小，<=0表示不使用chan
func (s *_TcpSession) RecvChanSize(size int) error {
	if s.state != gate.SessionState_Handshake {
		return errors.New("session state is not handshake")
	}

	if size <= 0 {
		s.recvChan = nil
		return nil
	}

	s.recvChan = make(chan gate.Recv, size)

	return nil
}

// RecvEventHandlers 设置接收自定义事件的处理器
func (s *_TcpSession) RecvEventHandlers(handlers []gate.RecvEventHandler) error {
	if s.state != gate.SessionState_Handshake {
		return errors.New("session state is not handshake")
	}

	s.dispatcher.EventHandlers = append(s.dispatcher.EventHandlers, handlers...)

	return nil
}

// RecvEventSize 设置自定义事件的chan的大小，<=0表示不使用chan
func (s *_TcpSession) RecvEventSize(size int) error {
	if s.state != gate.SessionState_Handshake {
		return errors.New("session state is not handshake")
	}

	if size <= 0 {
		s.recvEventChan = nil
		return nil
	}

	s.recvEventChan = make(chan gate.RecvEvent, size)

	return nil
}
