package gtp_gate

import (
	"context"
	"errors"
	"github.com/segmentio/ksuid"
	"kit.golaxy.org/plugins/gtp/codec"
	"kit.golaxy.org/plugins/gtp/transport"
	"net"
)

// _Acceptor 网络连接接受器
type _Acceptor struct {
	Gate    *_Gate
	Options *GateOptions
	encoder *codec.Encoder
	decoder *codec.Decoder
}

// accept 接受网络连接
func (acc *_Acceptor) accept(conn net.Conn) (*_Session, error) {
	return acc.handshake(conn)
}

// newSession 创建会话
func (acc *_Acceptor) newSession(conn net.Conn) (*_Session, error) {
	if conn == nil {
		return nil, errors.New("conn is nil")
	}

	session := &_Session{
		gate:  acc.Gate,
		id:    ksuid.New().String(),
		state: SessionState_Birth,
	}

	session.Context, session.cancel = context.WithCancel(acc.Gate.ctx)
	session.transceiver.Conn = conn

	// 初始化会话默认选项
	_SessionOption{}.Default()(&session.options)
	_SessionOption{}.SendDataChanSize(acc.Options.SessionSendDataChanSize)(&session.options)
	_SessionOption{}.RecvDataChanSize(acc.Options.SessionRecvDataChanSize)(&session.options)
	_SessionOption{}.SendEventChanSize(acc.Options.SessionSendEventSize)(&session.options)
	_SessionOption{}.RecvEventChanSize(acc.Options.SessionRecvEventSize)(&session.options)

	// 初始化消息事件分发器
	session.eventDispatcher.Transceiver = &session.transceiver
	session.eventDispatcher.RetryTimes = acc.Gate.options.IORetryTimes
	session.eventDispatcher.EventHandlers = []transport.EventHandler{session.trans.EventHandler, session.ctrl.EventHandler, session.handleEvent}

	// 初始化传输协议
	session.trans.Transceiver = &session.transceiver
	session.trans.RetryTimes = acc.Gate.options.IORetryTimes
	session.trans.PayloadHandler = session.handlePayload

	// 初始化控制协议
	session.ctrl.Transceiver = &session.transceiver
	session.ctrl.RetryTimes = acc.Gate.options.IORetryTimes
	session.ctrl.HeartbeatHandler = session.handleHeartbeat

	return session, nil
}
