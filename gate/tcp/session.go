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
	"sync"
)

// newTcpSession 创建会话
func newTcpSession(tcpGate *_TcpGate, conn net.Conn) *_TcpSession {
	session := &_TcpSession{
		gate:  tcpGate,
		id:    ksuid.New().String(),
		state: gate.SessionState_Birth,
	}

	session.Context, session.cancel = context.WithCancel(tcpGate.ctx)
	session.transceiver.Conn = conn

	// 初始化消息事件分发器
	session.dispatcher.Transceiver = &session.transceiver
	session.dispatcher.EventHandlers = []protocol.EventHandler{session.trans.EventHandler, session.ctrl.EventHandler, session.EventHandler}

	// 初始化传输协议
	session.trans.Transceiver = &session.transceiver
	session.trans.PayloadHandler = session.PayloadHandler

	// 初始化控制协议
	session.ctrl.Transceiver = &session.transceiver

	return session
}

type _TcpSession struct {
	context.Context
	sync.Mutex
	cancel               context.CancelFunc
	gate                 *_TcpGate
	id                   string
	token                string
	state                gate.SessionState
	transceiver          protocol.Transceiver
	dispatcher           protocol.EventDispatcher
	trans                protocol.TransProtocol
	ctrl                 protocol.CtrlProtocol
	stateChangedHandlers []gate.StateChangedHandler
	recvDataHandlers     []gate.RecvDataHandler
	recvEventHandlers    []gate.RecvEventHandler
	recvDataChan         chan gate.RecvData
	recvEventChan        chan gate.RecvEvent
}

// String implements fmt.Stringer
func (s *_TcpSession) String() string {
	return fmt.Sprintf("{Id:%s Token:%s State:%d}", s.GetId(), s.GetToken(), s.GetState())
}

// GetId 获取会话Id
func (s *_TcpSession) GetId() string {
	return s.id
}

// GetToken 获取token
func (s *_TcpSession) GetToken() string {
	return s.token
}

// GetState 获取会话状态
func (s *_TcpSession) GetState() gate.SessionState {
	s.Lock()
	defer s.Unlock()
	return s.state
}

// GetGroups 获取所属的会话组Id
func (s *_TcpSession) GetGroups() []string {
	return nil
}

// GetListenAddr 获取监听地址
func (s *_TcpSession) GetListenAddr() net.Addr {
	s.Lock()
	defer s.Unlock()

	if s.transceiver.Conn == nil {
		return nil
	}

	return s.transceiver.Conn.LocalAddr()
}

// GetClientAddr 获取客户端地址
func (s *_TcpSession) GetClientAddr() net.Addr {
	s.Lock()
	defer s.Unlock()

	if s.transceiver.Conn == nil {
		return nil
	}

	return s.transceiver.Conn.RemoteAddr()
}

// SendData 发送数据
func (s *_TcpSession) SendData(data []byte, sequenced bool) error {
	return s.trans.SendData(data, sequenced)
}

// SendEvent 发送自定义事件
func (s *_TcpSession) SendEvent(event protocol.Event[transport.Msg]) error {
	return protocol.Retry{
		Transceiver: &s.transceiver,
		Times:       s.gate.options.IORetryTimes,
	}.Send(s.transceiver.Send(event))
}

// RecvDataChan 接收数据的chan
func (s *_TcpSession) RecvDataChan() <-chan gate.RecvData {
	if s.recvDataChan == nil {
		ch := make(chan gate.RecvData, 1)
		ch <- gate.RecvData{Error: errors.New("RecvDataChan is not used")}
		close(ch)
		return ch
	}
	return s.recvDataChan
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
