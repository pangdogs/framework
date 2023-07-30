package tcp

import (
	"bytes"
	"errors"
	"golang.org/x/net/context"
	"kit.golaxy.org/golaxy/util"
	"kit.golaxy.org/plugins/gate"
	"kit.golaxy.org/plugins/internal"
	"kit.golaxy.org/plugins/logger"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/protocol"
	"net"
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
			logger.Errorf(s.gate.ctx, "session %q stopped", s.GetId())
		}
	}()

	s.SetState(gate.SessionState_Active)

	s.dispatcher.Run(s)

	s.SetState(gate.SessionState_Death)
}

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

// HeartbeatHandler Heartbeat消息事件处理器
func (s *_TcpSession) HeartbeatHandler(event protocol.Event[*transport.MsgHeartbeat]) error {
	return nil
}

// ErrorHandler 消息事件处理器
func (s *_TcpSession) ErrorHandler(ctx context.Context, err error) {

}
