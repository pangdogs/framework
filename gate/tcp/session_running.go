package tcp

import (
	"bytes"
	"errors"
	"kit.golaxy.org/golaxy/util"
	"kit.golaxy.org/plugins/gate"
	"kit.golaxy.org/plugins/internal"
	"kit.golaxy.org/plugins/logger"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/protocol"
	"net"
	"sync/atomic"
	"time"
)

// Init 初始化
func (s *_TcpSession) Init(transceiver *protocol.Transceiver, token string) {
	s.Lock()
	defer s.Unlock()

	// 初始化消息收发器
	s.transceiver.Conn = transceiver.Conn
	s.transceiver.Encoder = transceiver.Encoder
	s.transceiver.Decoder = transceiver.Decoder
	s.transceiver.Timeout = transceiver.Timeout
	s.transceiver.SequencedBuff.Reset(transceiver.SequencedBuff.SendSeq, transceiver.SequencedBuff.RecvSeq, transceiver.SequencedBuff.Cap)

	// 初始化token
	s.token = token
}

// Renew 刷新
func (s *_TcpSession) Renew(conn net.Conn, remoteRecvSeq uint32) (sendSeq, recvSeq uint32, err error) {
	s.Lock()
	defer s.Unlock()

	// 刷新链路
	sendSeq, recvSeq, err = s.transceiver.Renew(conn, remoteRecvSeq)
	if err != nil {
		return 0, 0, err
	}

	return s.transceiver.SequencedBuff.SendSeq, s.transceiver.SequencedBuff.RecvSeq, nil
}

// PauseIO 暂停收发消息
func (s *_TcpSession) PauseIO() {
	s.transceiver.Pause()
}

// ContinueIO 继续收发消息
func (s *_TcpSession) ContinueIO() {
	s.transceiver.Continue()
}

// Run 运行（会话的主线程）
func (s *_TcpSession) Run() {
	defer func() {
		if panicErr := util.Panic2Err(); panicErr != nil {
			logger.Errorf(s.gate.ctx, "session %q panicked, %s", s.GetId(), panicErr)
		}

		// 调整会话状态为已过期
		s.SetState(gate.SessionState_Death)

		// 关闭链接
		s.transceiver.Conn.Close()

		// 删除会话
		s.gate.sessionMap.Delete(s.GetId())
		atomic.AddInt64(&s.gate.sessionCount, -1)
	}()

	pinged := false
	var timeout time.Time

	// 调整会话状态为活跃
	s.SetState(gate.SessionState_Active)

	for {
		// 非活跃状态，检测超时时间
		if s.state == gate.SessionState_Inactive {
			if time.Now().After(timeout) {
				s.cancel()
			}
		}

		// 检测会话是否已关闭
		select {
		case <-s.Done():
			return
		default:
		}

		// 分发消息事件
		if err := s.dispatcher.Dispatching(); err != nil {
			// 网络io超时，触发心跳检测，向对方发送ping
			if errors.Is(err, protocol.ErrTimeout) {
				if !pinged {
					s.ctrl.SendPing()
					pinged = true
				} else {
					// 未收到对方回复pong或其他消息事件，再次网络io超时，调整会话状态不活跃
					if s.SetState(gate.SessionState_Inactive) {
						timeout = time.Now().Add(s.gate.options.SessionInactiveTimeout)
					}
				}
				continue
			}

			// 其他网络io类错误，调整会话状态不活跃
			if errors.Is(err, protocol.ErrNetIO) {
				if s.SetState(gate.SessionState_Inactive) {
					timeout = time.Now().Add(s.gate.options.SessionInactiveTimeout)
				}
				continue
			}

			logger.Debugf(s.gate.ctx, "session %q dispatching event failed, %s", s.GetId(), err)
		}

		// 没有错误，或非网络io类错误，重置ping状态
		pinged = false

		// 调整会话状态活跃
		s.SetState(gate.SessionState_Active)
	}
}

// SetState 调整会话状态
func (s *_TcpSession) SetState(state gate.SessionState) bool {
	old := s.state

	if old == state {
		return false
	}

	s.Lock()
	s.state = state
	s.Unlock()

	for i := range s.gate.options.SessionStateChangedHandlers {
		handler := s.gate.options.SessionStateChangedHandlers[i]
		if handler == nil {
			continue
		}
		err := internal.CallVoid(func() { handler(s, old, state) })
		if err != nil {
			logger.Errorf(s.gate.ctx, "session %q state changed handler error: %s", s.GetId(), err)
		}
	}

	for i := range s.stateChangedHandlers {
		handler := s.stateChangedHandlers[i]
		if handler == nil {
			continue
		}
		err := internal.CallVoid(func() { handler(s, old, state) })
		if err != nil {
			logger.Errorf(s.gate.ctx, "session %q state changed handler error: %s", s.GetId(), err)
		}
	}

	return true
}

// EventHandler 接收自定义事件的处理器
func (s *_TcpSession) EventHandler(event protocol.Event[transport.Msg]) error {
	if s.recvEventChan != nil {
		select {
		case s.recvEventChan <- gate.RecvEvent{Event: event.Clone()}:
		default:
			logger.Errorf(s.gate.ctx, "session %q RecvEventChan is full", s.GetId())
		}
	}

	for i := range s.gate.options.SessionRecvEventHandlers {
		handler := s.gate.options.SessionRecvEventHandlers[i]
		if handler == nil {
			continue
		}
		err := internal.Call(func() error { return handler(s, event) })
		if err == nil || !errors.As(err, protocol.ErrUnexpectedMsg) {
			return err
		}
	}

	for i := range s.recvEventHandlers {
		handler := s.recvEventHandlers[i]
		if handler == nil {
			continue
		}
		err := internal.Call(func() error { return handler(s, event) })
		if err == nil || !errors.As(err, protocol.ErrUnexpectedMsg) {
			return err
		}
	}

	return protocol.ErrUnexpectedMsg
}

// PayloadHandler Payload消息事件处理器
func (s *_TcpSession) PayloadHandler(event protocol.Event[*transport.MsgPayload]) error {
	if s.recvDataChan != nil {
		select {
		case s.recvDataChan <- gate.RecvData{
			Data:      bytes.Clone(event.Msg.Data),
			Sequenced: event.Flags.Is(transport.Flag_Sequenced),
		}:
		default:
			logger.Errorf(s.gate.ctx, "session %q RecvChan is full", s.GetId())
		}
	}

	for i := range s.gate.options.SessionRecvDataHandlers {
		handler := s.gate.options.SessionRecvDataHandlers[i]
		if handler == nil {
			continue
		}
		err := internal.Call(func() error { return handler(s, event.Msg.Data, event.Flags.Is(transport.Flag_Sequenced)) })
		if err == nil || !errors.As(err, protocol.ErrUnexpectedMsg) {
			return err
		}
	}

	for i := range s.recvDataHandlers {
		handler := s.recvDataHandlers[i]
		if handler == nil {
			continue
		}
		err := internal.Call(func() error { return handler(s, event.Msg.Data, event.Flags.Is(transport.Flag_Sequenced)) })
		if err == nil || !errors.As(err, protocol.ErrUnexpectedMsg) {
			return err
		}
	}

	return protocol.ErrUnexpectedMsg
}
