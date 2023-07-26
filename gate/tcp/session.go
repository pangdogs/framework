package tcp

import (
	"errors"
	"fmt"
	"github.com/segmentio/ksuid"
	"golang.org/x/net/context"
	"kit.golaxy.org/plugins/gate"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/protocol"
	"net"
)

func newTcpSession(tcpGate *_TcpGate) *_TcpSession {
	session := &_TcpSession{
		gate:  tcpGate,
		id:    ksuid.New().String(),
		state: gate.SessionState_Birth,
	}

	session.Context, session.cancel = context.WithCancel(tcpGate.ctx)

	// 初始化消息事件分发器
	session.dispatcher.Transceiver = &session.transceiver
	session.dispatcher.EventHandlers = []protocol.EventHandler{session.trans.EventHandler, session.ctrl.EventHandler}

	for i := range session.gate.options.SessionRecvEventHandlers {
		handler := session.gate.options.SessionRecvEventHandlers[i]
		if handler == nil {
			continue
		}
		session.dispatcher.EventHandlers = append(session.dispatcher.EventHandlers, func(event protocol.Event[transport.Msg]) error { return handler(session, event) })
	}

	session.dispatcher.EventHandlers = append(session.dispatcher.EventHandlers, session.EventHandler)
	session.dispatcher.ErrorHandler = session.ErrorHandler

	// 初始化传输协议
	session.trans.Transceiver = &session.transceiver
	session.trans.PayloadHandler = session.PayloadHandler

	// 初始化控制协议
	session.ctrl.Transceiver = &session.transceiver
	session.ctrl.HeartbeatHandler = session.HeartbeatHandler

	return session
}

type _TcpSession struct {
	context.Context
	gate                *_TcpGate
	cancel              context.CancelFunc
	id                  string
	state               gate.SessionState
	token               string
	transceiver         protocol.Transceiver
	dispatcher          protocol.EventDispatcher
	trans               protocol.TransProtocol
	ctrl                protocol.CtrlProtocol
	stateChangedHandler gate.StateChangedHandler
	recvHandler         gate.RecvHandler
	recvChan            chan gate.Recv
	recvEventChan       chan gate.RecvEvent
}

// String implements fmt.Stringer
func (s *_TcpSession) String() string {
	return fmt.Sprintf("{Id:%s Token:%s State:%d}", s.GetId(), s.GetToken(), s.GetState())
}

// GetId 获取会话Id
func (s *_TcpSession) GetId() string {
	return s.id
}

// GetState 获取会话状态
func (s *_TcpSession) GetState() gate.SessionState {
	return s.state
}

// GetToken 获取token
func (s *_TcpSession) GetToken() string {
	return s.token
}

// GetGroups 获取所属的会话组Id
func (s *_TcpSession) GetGroups() []string {
	return nil
}

// GetListenAddr 获取监听地址
func (s *_TcpSession) GetListenAddr() net.Addr {
	return s.transceiver.Conn.LocalAddr()
}

// GetClientAddr 获取客户端地址
func (s *_TcpSession) GetClientAddr() net.Addr {
	return s.transceiver.Conn.RemoteAddr()
}

// Send 发送数据
func (s *_TcpSession) Send(data []byte, sequenced bool) error {
	return s.trans.SendData(data, sequenced)
}

// RecvChan 接收数据的chan
func (s *_TcpSession) RecvChan() <-chan gate.Recv {
	if s.recvChan == nil {
		ch := make(chan gate.Recv, 1)
		ch <- gate.Recv{Error: errors.New("RecvChan is not used")}
		close(ch)
		return ch
	}
	return s.recvChan
}

// SendEvent 发送自定义事件
func (s *_TcpSession) SendEvent(event protocol.Event[transport.Msg]) error {
	return protocol.Retry{
		Transceiver: &s.transceiver,
		Times:       s.gate.options.IORetryTimes,
	}.Send(s.transceiver.Send(event))
}

// RecvEventChan 接收自定义事件的chan
func (s *_TcpSession) RecvEventChan() <-chan gate.RecvEvent {
	if s.recvEventChan == nil {
		ch := make(chan gate.RecvEvent, 1)
		ch <- gate.RecvEvent{Error: errors.New("RecvEventChan is not used")}
		close(ch)
		return ch
	}
	return s.recvEventChan
}

// Close 关闭连接
func (s *_TcpSession) Close(err error) {
	if err != nil {
		s.ctrl.SendRst(err)
	}
	s.cancel()
}
