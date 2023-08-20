package gtp

import (
	"errors"
	"github.com/segmentio/ksuid"
	"golang.org/x/net/context"
	"kit.golaxy.org/plugins/gate"
	"kit.golaxy.org/plugins/transport/codec"
	"kit.golaxy.org/plugins/transport/protocol"
	"net"
)

// _Acceptor 网络连接接受器
type _Acceptor struct {
	Gate    *_GtpGate
	Options *GateOptions
	encoder *codec.Encoder
	decoder *codec.Decoder
}

// Accept 接受网络连接
func (acc *_Acceptor) Accept(conn net.Conn) (*_GtpSession, error) {
	return acc.handshake(conn)
}

// newGtpSession 创建会话
func (acc *_Acceptor) newGtpSession(conn net.Conn) (*_GtpSession, error) {
	if conn == nil {
		return nil, errors.New("conn is nil")
	}

	session := &_GtpSession{
		gate:  acc.Gate,
		id:    ksuid.New().String(),
		state: gate.SessionState_Birth,
	}

	session.Context, session.cancel = context.WithCancel(acc.Gate.ctx)
	session.transceiver.Conn = conn

	// 初始化消息事件分发器
	session.dispatcher.Transceiver = &session.transceiver
	session.dispatcher.RetryTimes = acc.Gate.options.IORetryTimes
	session.dispatcher.EventHandlers = []protocol.EventHandler{session.trans.EventHandler, session.ctrl.EventHandler, session.EventHandler}

	// 初始化传输协议
	session.trans.Transceiver = &session.transceiver
	session.trans.RetryTimes = acc.Gate.options.IORetryTimes
	session.trans.PayloadHandler = session.PayloadHandler

	// 初始化控制协议
	session.ctrl.Transceiver = &session.transceiver
	session.ctrl.RetryTimes = acc.Gate.options.IORetryTimes
	session.ctrl.HeartbeatHandler = session.HeartbeatHandler

	return session, nil
}
