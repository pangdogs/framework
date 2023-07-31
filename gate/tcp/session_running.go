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
)

// Init 初始化
func (s *_TcpSession) Init(transceiver protocol.Transceiver, token string) {
	s.transceiver = transceiver
	s.token = token
}

// Renew 更新
func (s *_TcpSession) Renew(conn net.Conn, remoteRecvSeq uint32) (sendSeq, recvSeq uint32, err error) {
	// 切换连接
	s.transceiver.Conn.Close()
	s.transceiver.Conn = conn

	// 同步对端时序
	if !s.transceiver.SequencedBuff.Synchronization(remoteRecvSeq) {
		return 0, 0, errors.New("io sequenced buff synchronization failed")
	}

	return s.transceiver.SequencedBuff.SendSeq, s.transceiver.SequencedBuff.RecvSeq, nil
}

// Run 运行（会话的主线程）
func (s *_TcpSession) Run() {
	defer func() {
		if panicErr := util.Panic2Err(); panicErr != nil {
			logger.Errorf(s.gate.ctx, "session %q panicked, %s", s.GetId(), panicErr)
		}
		
		// 调整会话状态为已过期
		s.SetState(gate.SessionState_Death)

		// 删除会话
		s.gate.sessionMap.Delete(s.GetId())
		atomic.AddInt64(&s.gate.sessionCount, -1)
	}()

	pinged := false

	// 调整会话状态为活跃
	s.SetState(gate.SessionState_Active)

	for {
		select {
		case <-s.Done():
			return
		default:
		}

		// 分发消息事件
		if err := s.dispatcher.Dispatching(); err != nil {
			// 网络io超时，触发心跳检测，向对方发送ping
			if errors.Is(err, protocol.ErrTimeout) {
				if pinged {
					// 未收到对方回复pong或其他消息事件，再次网络io超时，调整会话状态不活跃
					s.SetState(gate.SessionState_Inactive)
				} else {
					s.ctrl.SendPing()
					pinged = true
				}
				continue
			}

			// 其他网络io类错误，调整会话状态不活跃
			if errors.Is(err, protocol.ErrNetIO) {
				s.SetState(gate.SessionState_Inactive)
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
func (s *_TcpSession) SetState(state gate.SessionState) {
	old := s.state

	if old == state {
		return
	}
	s.state = state

	var session gate.Session

	switch s.state {
	case gate.SessionState_Handshake:
		session = &_TcpSessionHandshake{_TcpSession: s}
	default:
		session = s
	}

	if s.gate.options.SessionStateChangedHandler != nil {
		internal.CallVoid(func() { s.gate.options.SessionStateChangedHandler(session, old, state) })
	}

	if s.stateChangedHandler != nil {
		internal.CallVoid(func() { s.stateChangedHandler(old, state) })
	}
}

// EventHandler 接收自定义事件的处理器
func (s *_TcpSession) EventHandler(event protocol.Event[transport.Msg]) error {
	if s.recvEventChan != nil {
		recvEvent := gate.RecvEvent{}
		recvEvent.Event = event.Clone()

		select {
		case s.recvEventChan <- recvEvent:
		default:
			logger.Errorf(s.gate.ctx, "session %q RecvEventChan is full", s.GetId())
		}
	}
	return protocol.ErrUnexpectedMsg
}

// PayloadHandler Payload消息事件处理器
func (s *_TcpSession) PayloadHandler(event protocol.Event[*transport.MsgPayload]) error {
	if s.recvChan != nil {
		select {
		case s.recvChan <- gate.Recv{
			Data:      bytes.Clone(event.Msg.Data),
			Sequenced: event.Flags.Is(transport.Flag_Sequenced),
		}:
		default:
			logger.Errorf(s.gate.ctx, "session %q RecvChan is full", s.GetId())
		}
	}

	if s.gate.options.SessionRecvHandler != nil {
		err := internal.Call(func() error {
			return s.gate.options.SessionRecvHandler(s, event.Msg.Data, event.Flags.Is(transport.Flag_Sequenced))
		})
		if err == nil || !errors.As(err, protocol.ErrUnexpectedMsg) {
			return err
		}
	}

	if s.recvHandler == nil {
		return protocol.ErrUnexpectedMsg
	}

	return s.recvHandler(event.Msg.Data, event.Flags.Is(transport.Flag_Sequenced))
}
