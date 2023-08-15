package gtp

import (
	"errors"
	"fmt"
	"github.com/segmentio/ksuid"
	"golang.org/x/net/context"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/plugins/gate"
	"kit.golaxy.org/plugins/logger"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/protocol"
	"net"
	"sync"
)

// newGtpSession 创建会话
func newGtpSession(gtpGate *_GtpGate, conn net.Conn) (*_GtpSession, error) {
	if conn == nil {
		return nil, errors.New("conn is nil")
	}

	session := &_GtpSession{
		gate:  gtpGate,
		id:    ksuid.New().String(),
		state: gate.SessionState_Birth,
	}

	session.Context, session.cancel = context.WithCancel(gtpGate.ctx)
	session.transceiver.Conn = conn

	// 初始化消息事件分发器
	session.dispatcher.Transceiver = &session.transceiver
	session.dispatcher.RetryTimes = gtpGate.options.IORetryTimes
	session.dispatcher.EventHandlers = []protocol.EventHandler{session.trans.EventHandler, session.ctrl.EventHandler, session.EventHandler}

	// 初始化传输协议
	session.trans.Transceiver = &session.transceiver
	session.trans.RetryTimes = gtpGate.options.IORetryTimes
	session.trans.PayloadHandler = session.PayloadHandler

	// 初始化控制协议
	session.ctrl.Transceiver = &session.transceiver
	session.ctrl.RetryTimes = gtpGate.options.IORetryTimes

	return session, nil
}

type _GtpSession struct {
	context.Context
	sync.Mutex
	cancel               context.CancelFunc
	gate                 *_GtpGate
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
	sendDataChan         chan gate.SendData
	recvDataChan         chan gate.RecvData
	sendEventChan        chan protocol.Event[transport.Msg]
	recvEventChan        chan gate.RecvEvent
}

// String implements fmt.Stringer
func (s *_GtpSession) String() string {
	return fmt.Sprintf("{Id:%s Token:%s State:%d}", s.GetId(), s.GetToken(), s.GetState())
}

// GetContext 获取服务上下文
func (s *_GtpSession) GetContext() service.Context {
	return s.gate.ctx
}

// GetId 获取会话Id
func (s *_GtpSession) GetId() string {
	return s.id
}

// GetToken 获取token
func (s *_GtpSession) GetToken() string {
	return s.token
}

// GetState 获取会话状态
func (s *_GtpSession) GetState() gate.SessionState {
	s.Lock()
	defer s.Unlock()
	return s.state
}

// GetGroups 获取所属的会话组Id
func (s *_GtpSession) GetGroups() []string {
	return nil
}

// GetListenAddr 获取监听地址
func (s *_GtpSession) GetListenAddr() net.Addr {
	s.Lock()
	defer s.Unlock()
	return s.transceiver.Conn.LocalAddr()
}

// GetClientAddr 获取客户端地址
func (s *_GtpSession) GetClientAddr() net.Addr {
	s.Lock()
	defer s.Unlock()
	return s.transceiver.Conn.RemoteAddr()
}

// SendData 发送数据
func (s *_GtpSession) SendData(data []byte, sequenced bool) error {
	return s.trans.SendData(data, sequenced)
}

// SendEvent 发送自定义事件
func (s *_GtpSession) SendEvent(event protocol.Event[transport.Msg]) error {
	return protocol.Retry{
		Transceiver: &s.transceiver,
		Times:       s.gate.options.IORetryTimes,
	}.Send(s.transceiver.Send(event))
}

// SendDataChan 发送数据的channel
func (s *_GtpSession) SendDataChan() chan<- gate.SendData {
	if s.sendDataChan == nil {
		logger.Panicf(s.gate.ctx, "send data channel size less equal 0, can't be used")
	}
	return s.sendDataChan
}

// RecvDataChan 接收数据的channel
func (s *_GtpSession) RecvDataChan() <-chan gate.RecvData {
	if s.recvDataChan == nil {
		logger.Panicf(s.gate.ctx, "receive data channel size less equal 0, can't be used")
	}
	return s.recvDataChan
}

// SendEventChan 发送自定义事件的channel
func (s *_GtpSession) SendEventChan() chan<- protocol.Event[transport.Msg] {
	if s.sendEventChan == nil {
		logger.Panicf(s.gate.ctx, "send event channel size less equal 0, can't be used")
	}
	return s.sendEventChan
}

// RecvEventChan 接收自定义事件的channel
func (s *_GtpSession) RecvEventChan() <-chan gate.RecvEvent {
	if s.recvEventChan == nil {
		logger.Panicf(s.gate.ctx, "receive event channel size less equal 0, can't be used")
	}
	return s.recvEventChan
}

// Close 关闭
func (s *_GtpSession) Close(err error) {
	if err != nil {
		s.ctrl.SendRst(err)
	}
	s.cancel()
}
