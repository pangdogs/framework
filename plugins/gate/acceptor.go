package gate

import (
	"context"
	"errors"
	"git.golaxy.org/core/util/generic"
	"git.golaxy.org/core/util/uid"
	"git.golaxy.org/framework/net/gtp/codec"
	"git.golaxy.org/framework/util/concurrent"
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

	ctx, _ := context.WithTimeout(acc.gate.ctx, acc.gate.options.AcceptTimeout)

	return acc.handshake(ctx, conn)
}

// newSession 创建会话
func (acc *_Acceptor) newSession(conn net.Conn) (*_Session, error) {
	if conn == nil {
		return nil, errors.New("conn is nil")
	}

	session := &_Session{
		terminatedChan: make(chan struct{}),
		gate:           acc.gate,
		id:             uid.New(),
		state:          SessionState_Birth,
	}

	session.Context, session.terminate = context.WithCancelCause(acc.gate.ctx)
	session.transceiver.Conn = conn

	// 初始化会话默认选项
	sessionWith.Default()(&session.options)
	sessionWith.SendDataChanSize(acc.gate.options.SessionSendDataChanSize)(&session.options)
	sessionWith.RecvDataChanSize(acc.gate.options.SessionRecvDataChanSize, acc.gate.options.SessionRecvDataChanRecyclable)(&session.options)
	sessionWith.SendEventChanSize(acc.gate.options.SessionSendEventChanSize)(&session.options)
	sessionWith.RecvEventChanSize(acc.gate.options.SessionRecvEventChanSize)(&session.options)

	// 初始化消息事件分发器
	session.eventDispatcher.Transceiver = &session.transceiver
	session.eventDispatcher.RetryTimes = acc.gate.options.IORetryTimes
	session.eventDispatcher.EventHandler = generic.MakeDelegateFunc1(session.trans.HandleEvent, session.ctrl.HandleEvent, session.handleRecvEventChan, session.handleRecvEvent)

	// 初始化传输协议
	session.trans.Transceiver = &session.transceiver
	session.trans.RetryTimes = acc.gate.options.IORetryTimes
	session.trans.PayloadHandler = generic.MakeDelegateFunc1(session.handleRecvDataChan, session.handleRecvPayload)

	// 初始化控制协议
	session.ctrl.Transceiver = &session.transceiver
	session.ctrl.RetryTimes = acc.gate.options.IORetryTimes
	session.ctrl.HeartbeatHandler = generic.MakeDelegateFunc1(session.handleRecvHeartbeat)

	// 初始化监听器
	session.dataWatchers = concurrent.MakeLockedSlice[*_DataWatcher](0, 0)
	session.eventWatchers = concurrent.MakeLockedSlice[*_EventWatcher](0, 0)

	return session, nil
}
