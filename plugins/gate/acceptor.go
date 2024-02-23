package gate

import (
	"context"
	"errors"
	"git.golaxy.org/core/util/generic"
	"git.golaxy.org/framework/net/gtp/codec"
	"git.golaxy.org/framework/util/concurrent"
	"github.com/segmentio/ksuid"
	"net"
)

// _Acceptor 网络连接接受器
type _Acceptor struct {
	gate           *_Gate
	encoderCreator codec.EncoderCreator
	decoderCreator codec.DecoderCreator
}

// accept 接受网络连接
func (acc *_Acceptor) accept(conn net.Conn) (*_Session, error) {
	select {
	case <-acc.gate.ctx.Done():
		return nil, errors.New("service shutdown")
	default:
	}

	return acc.handshake(conn)
}

// newSession 创建会话
func (acc *_Acceptor) newSession(conn net.Conn) (*_Session, error) {
	if conn == nil {
		return nil, errors.New("conn is nil")
	}

	session := &_Session{
		gate:  acc.gate,
		id:    ksuid.New().String(),
		state: SessionState_Birth,
	}

	session.Context, session.cancel = context.WithCancelCause(acc.gate.ctx)
	session.transceiver.Conn = conn

	// 初始化会话默认选项
	sessionWith.Default()(&session.options)
	sessionWith.SendDataChanSize(acc.gate.options.SessionSendDataChanSize)(&session.options)
	sessionWith.RecvDataChanSize(acc.gate.options.SessionRecvDataChanSize)(&session.options)
	sessionWith.SendEventChanSize(acc.gate.options.SessionSendEventChanSize)(&session.options)
	sessionWith.RecvEventChanSize(acc.gate.options.SessionRecvEventChanSize)(&session.options)

	// 初始化消息事件分发器
	session.eventDispatcher.Transceiver = &session.transceiver
	session.eventDispatcher.RetryTimes = acc.gate.options.IORetryTimes
	session.eventDispatcher.EventHandler = generic.CastDelegateFunc1(session.trans.HandleEvent, session.ctrl.HandleEvent, session.handleRecvEventChan, session.handleEventProcess)

	// 初始化传输协议
	session.trans.Transceiver = &session.transceiver
	session.trans.RetryTimes = acc.gate.options.IORetryTimes
	session.trans.PayloadHandler = generic.CastDelegateFunc1(session.handleRecvDataChan, session.handlePayloadProcess)

	// 初始化控制协议
	session.ctrl.Transceiver = &session.transceiver
	session.ctrl.RetryTimes = acc.gate.options.IORetryTimes
	session.ctrl.HeartbeatHandler = generic.CastDelegateFunc1(session.handleHeartbeat)

	// 初始化监听器
	session.dataWatchers = concurrent.MakeLockedSlice[*_DataWatcher](0, 0)
	session.eventWatchers = concurrent.MakeLockedSlice[*_EventWatcher](0, 0)

	return session, nil
}
