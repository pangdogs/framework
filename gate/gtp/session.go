package gtp

import (
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/plugins/gate"
	"kit.golaxy.org/plugins/logger"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/protocol"
	"net"
	"sync"
)

type _GtpSession struct {
	context.Context
	sync.Mutex
	cancel      context.CancelFunc
	gate        *_GtpGate
	options     gate.SessionOptions
	id          string
	token       string
	state       gate.SessionState
	transceiver protocol.Transceiver
	dispatcher  protocol.EventDispatcher
	trans       protocol.TransProtocol
	ctrl        protocol.CtrlProtocol
	renewChan   chan struct{}
}

// String implements fmt.Stringer
func (s *_GtpSession) String() string {
	return fmt.Sprintf("{Id:%s Token:%s State:%d}", s.GetId(), s.GetToken(), s.GetState())
}

// Options 设置会话选项（在会话状态Handshake与Confirmed时可用）
func (s *_GtpSession) Options(options ...gate.SessionOption) error {
	s.Lock()
	defer s.Unlock()

	switch s.state {
	case gate.SessionState_Handshake, gate.SessionState_Confirmed:
		break
	default:
		return errors.New("incorrect session state")
	}

	gate.Option{}.Default()(&s.options)

	for i := range options {
		options[i](&s.options)
	}

	return nil
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

// GetLocalAddr 获取本地地址
func (s *_GtpSession) GetLocalAddr() net.Addr {
	s.Lock()
	defer s.Unlock()
	return s.transceiver.Conn.LocalAddr()
}

// GetRemoteAddr 获取对端地址
func (s *_GtpSession) GetRemoteAddr() net.Addr {
	s.Lock()
	defer s.Unlock()
	return s.transceiver.Conn.RemoteAddr()
}

// SendData 发送数据
func (s *_GtpSession) SendData(data []byte) error {
	return s.trans.SendData(data)
}

// SendEvent 发送自定义事件
func (s *_GtpSession) SendEvent(event protocol.Event[transport.Msg]) error {
	return protocol.Retry{
		Transceiver: &s.transceiver,
		Times:       s.gate.options.IORetryTimes,
	}.Send(s.transceiver.Send(event))
}

// SendDataChan 发送数据的channel
func (s *_GtpSession) SendDataChan() chan<- []byte {
	if s.options.SendDataChan == nil {
		logger.Panicf(s.gate.ctx, "send data channel size less equal 0, can't be used")
	}
	return s.options.SendDataChan
}

// RecvDataChan 接收数据的channel
func (s *_GtpSession) RecvDataChan() <-chan []byte {
	if s.options.RecvDataChan == nil {
		logger.Panicf(s.gate.ctx, "receive data channel size less equal 0, can't be used")
	}
	return s.options.RecvDataChan
}

// SendEventChan 发送自定义事件的channel
func (s *_GtpSession) SendEventChan() chan<- protocol.Event[transport.Msg] {
	if s.options.SendEventChan == nil {
		logger.Panicf(s.gate.ctx, "send event channel size less equal 0, can't be used")
	}
	return s.options.SendEventChan
}

// RecvEventChan 接收自定义事件的channel
func (s *_GtpSession) RecvEventChan() <-chan protocol.Event[transport.Msg] {
	if s.options.RecvEventChan == nil {
		logger.Panicf(s.gate.ctx, "receive event channel size less equal 0, can't be used")
	}
	return s.options.RecvEventChan
}

// Close 关闭
func (s *_GtpSession) Close(err error) {
	if err != nil {
		s.ctrl.SendRst(err)
	}
	s.cancel()
}
