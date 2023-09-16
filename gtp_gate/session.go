//go:generate stringer -type SessionState
package gtp_gate

import (
	"context"
	"errors"
	"fmt"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/plugins/gtp"
	"kit.golaxy.org/plugins/gtp/transport"
	"kit.golaxy.org/plugins/logger"
	"net"
	"sync"
)

// SessionState 客户端会话状态
type SessionState int32

const (
	SessionState_Birth     SessionState = iota // 出生
	SessionState_Handshake                     // 与客户端握手中
	SessionState_Confirmed                     // 已确认客户端连接
	SessionState_Active                        // 客户端活跃
	SessionState_Inactive                      // 客户端不活跃，等待重连恢复中
	SessionState_Death                         // 已过期
)

// Session 客户端会话
type Session interface {
	context.Context
	fmt.Stringer
	// Options 设置会话选项（在会话状态Handshake与Confirmed时可用）
	Options(options ...SessionOption) error
	// GetContext 获取服务上下文
	GetContext() service.Context
	// GetId 获取会话Id
	GetId() string
	// GetToken 获取token
	GetToken() string
	// GetState 获取会话状态
	GetState() SessionState
	// GetLocalAddr 获取本地地址
	GetLocalAddr() net.Addr
	// GetRemoteAddr 获取对端地址
	GetRemoteAddr() net.Addr
	// SendData 发送数据
	SendData(data []byte) error
	// SendEvent 发送自定义事件
	SendEvent(event transport.Event[gtp.Msg]) error
	// SendDataChan 发送数据的channel
	SendDataChan() chan<- []byte
	// RecvDataChan 接收数据的channel
	RecvDataChan() <-chan []byte
	// SendEventChan 发送自定义事件的channel
	SendEventChan() chan<- transport.Event[gtp.Msg]
	// RecvEventChan 接收自定义事件的channel
	RecvEventChan() <-chan transport.Event[gtp.Msg]
	// GetFutures 获取异步模型Future控制器
	GetFutures() transport.IFutures
	// Close 关闭
	Close(err error)
}

type _GtpSession struct {
	context.Context
	sync.Mutex
	cancel          context.CancelFunc
	gate            *_GtpGate
	options         SessionOptions
	id              string
	token           string
	state           SessionState
	transceiver     transport.Transceiver
	eventDispatcher transport.EventDispatcher
	trans           transport.TransProtocol
	ctrl            transport.CtrlProtocol
	renewChan       chan struct{}
}

// String implements fmt.Stringer
func (s *_GtpSession) String() string {
	return fmt.Sprintf("{Id:%s Token:%s State:%d}", s.GetId(), s.GetToken(), s.GetState())
}

// Options 设置会话选项（在会话状态Handshake与Confirmed时可用）
func (s *_GtpSession) Options(options ...SessionOption) error {
	s.Lock()
	defer s.Unlock()

	switch s.state {
	case SessionState_Handshake, SessionState_Confirmed:
		break
	default:
		return errors.New("incorrect session state")
	}

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
func (s *_GtpSession) GetState() SessionState {
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
func (s *_GtpSession) SendEvent(event transport.Event[gtp.Msg]) error {
	return transport.Retry{
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
func (s *_GtpSession) SendEventChan() chan<- transport.Event[gtp.Msg] {
	if s.options.SendEventChan == nil {
		logger.Panicf(s.gate.ctx, "send event channel size less equal 0, can't be used")
	}
	return s.options.SendEventChan
}

// RecvEventChan 接收自定义事件的channel
func (s *_GtpSession) RecvEventChan() <-chan transport.Event[gtp.Msg] {
	if s.options.RecvEventChan == nil {
		logger.Panicf(s.gate.ctx, "receive event channel size less equal 0, can't be used")
	}
	return s.options.RecvEventChan
}

// GetFutures 获取异步模型Future控制器
func (s *_GtpSession) GetFutures() transport.IFutures {
	return &s.gate.futures
}

// Close 关闭
func (s *_GtpSession) Close(err error) {
	if err != nil {
		s.ctrl.SendRst(err)
	}
	s.cancel()
}
