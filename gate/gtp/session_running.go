package gtp

import (
	"bytes"
	"errors"
	"fmt"
	"kit.golaxy.org/golaxy/util"
	"kit.golaxy.org/plugins/gate"
	"kit.golaxy.org/plugins/internal"
	"kit.golaxy.org/plugins/logger"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/codec"
	"kit.golaxy.org/plugins/transport/protocol"
	"math/rand"
	"net"
	"sync/atomic"
	"time"
)

// Init 初始化
func (s *_GtpSession) Init(conn net.Conn, encoder codec.IEncoder, decoder codec.IDecoder, token string) (sendSeq, recvSeq uint32) {
	s.Lock()
	defer s.Unlock()

	// 初始化消息收发器
	s.transceiver.Conn = conn
	s.transceiver.Encoder = encoder
	s.transceiver.Decoder = decoder
	s.transceiver.Timeout = s.gate.options.IOTimeout

	buff := &protocol.SequencedBuffer{}
	buff.Reset(rand.Uint32(), rand.Uint32(), s.gate.options.IOBufferCap)

	s.transceiver.Buffer = buff

	// 初始化token
	s.token = token

	// 初始化channel
	if s.gate.options.SessionSendDataChanSize > 0 {
		s.sendDataChan = make(chan []byte, s.gate.options.SessionSendDataChanSize)
	}
	if s.gate.options.SessionRecvDataChanSize > 0 {
		s.recvDataChan = make(chan []byte, s.gate.options.SessionRecvDataChanSize)
	}
	if s.gate.options.SessionSendEventSize > 0 {
		s.sendEventChan = make(chan protocol.Event[transport.Msg], s.gate.options.SessionSendEventSize)
	}
	if s.gate.options.SessionRecvEventSize > 0 {
		s.recvEventChan = make(chan protocol.Event[transport.Msg], s.gate.options.SessionRecvEventSize)
	}

	return buff.SendSeq(), buff.RecvSeq()
}

// Renew 刷新
func (s *_GtpSession) Renew(conn net.Conn, remoteRecvSeq uint32) (sendSeq, recvSeq uint32, err error) {
	s.Lock()
	defer s.Unlock()

	// 刷新链路
	return s.transceiver.Renew(conn, remoteRecvSeq)
}

// PauseIO 暂停收发消息
func (s *_GtpSession) PauseIO() {
	s.transceiver.Pause()
}

// ContinueIO 继续收发消息
func (s *_GtpSession) ContinueIO() {
	s.transceiver.Continue()
}

// Run 运行（会话的主线程）
func (s *_GtpSession) Run() {
	defer func() {
		if panicErr := util.Panic2Err(recover()); panicErr != nil {
			logger.Errorf(s.gate.ctx, "session %q panicked, %s", s.GetId(), fmt.Errorf("panicked: %w", panicErr))
		}

		// 调整会话状态为已过期
		s.SetState(gate.SessionState_Death)

		// 关闭连接和清理数据
		if s.transceiver.Conn != nil {
			s.transceiver.Conn.Close()
		}
		s.transceiver.Clean()

		// 删除会话
		s.gate.sessionMap.Delete(s.GetId())
		atomic.AddInt64(&s.gate.sessionCount, -1)
	}()

	pinged := false
	var timeout time.Time

	// 调整会话状态为活跃
	s.SetState(gate.SessionState_Active)

	// 启动发送数据的线程
	if s.sendDataChan != nil {
		go func() {
			for {
				select {
				case data := <-s.sendDataChan:
					if err := s.SendData(data); err != nil {
						logger.Errorf(s.gate.ctx, "session %q fetch data from the send data channel for sending failed, %s", s.GetId(), err)
					}
				case <-s.Done():
					return
				}
			}
		}()
	}

	// 启动发送自定义事件的线程
	if s.sendEventChan != nil {
		go func() {
			for {
				select {
				case event := <-s.sendEventChan:
					if err := s.SendEvent(event); err != nil {
						logger.Errorf(s.gate.ctx, "session %q fetch event from the send event channel for sending failed, %s", s.GetId(), err)
					}
				case <-s.Done():
					return
				}
			}
		}()
	}

	for {
		// 非活跃状态，检测超时时间
		if s.state == gate.SessionState_Inactive {
			if time.Now().After(timeout) {
				s.cancel()
			}
		}

		// 检测会话是否已关闭
		select {
		case <-s.Done():
			return
		default:
		}

		// 分发消息事件
		if err := s.dispatcher.Dispatching(); err != nil {
			// 网络io超时，触发心跳检测，向对方发送ping
			if errors.Is(err, protocol.ErrTimeout) {
				if !pinged {
					s.ctrl.SendPing()
					pinged = true
				} else {
					// 未收到对方回复pong或其他消息事件，再次网络io超时，调整会话状态不活跃
					if s.SetState(gate.SessionState_Inactive) {
						timeout = time.Now().Add(s.gate.options.SessionInactiveTimeout)
					}
				}
				continue
			}

			// 其他网络io类错误，调整会话状态不活跃
			if errors.Is(err, protocol.ErrNetIO) {
				if s.SetState(gate.SessionState_Inactive) {
					timeout = time.Now().Add(s.gate.options.SessionInactiveTimeout)
				}
				continue
			}

			logger.Debugf(s.gate.ctx, "session %q dispatching event failed, %s", s.GetId(), err)
			continue
		}

		// 没有错误，或非网络io类错误，重置ping状态
		pinged = false

		// 调整会话状态活跃
		if s.SetState(gate.SessionState_Active) {
			// 发送缓存的消息
			protocol.Retry{
				Transceiver: &s.transceiver,
				Times:       s.gate.options.IORetryTimes,
			}.Send(s.transceiver.Resend())
		}
	}
}

// SetState 调整会话状态
func (s *_GtpSession) SetState(state gate.SessionState) bool {
	old := s.state

	if old == state {
		return false
	}

	s.Lock()
	s.state = state
	s.Unlock()

	for i := range s.gate.options.SessionStateChangedHandlers {
		handler := s.gate.options.SessionStateChangedHandlers[i]
		if handler == nil {
			continue
		}
		err := internal.CallVoid(func() { handler(s, old, state) })
		if err != nil {
			logger.Errorf(s.gate.ctx, "session %q state changed handler error: %s", s.GetId(), err)
		}
	}

	for i := range s.stateChangedHandlers {
		handler := s.stateChangedHandlers[i]
		if handler == nil {
			continue
		}
		err := internal.CallVoid(func() { handler(s, old, state) })
		if err != nil {
			logger.Errorf(s.gate.ctx, "session %q state changed handler error: %s", s.GetId(), err)
		}
	}

	return true
}

// EventHandler 接收自定义事件的处理器
func (s *_GtpSession) EventHandler(event protocol.Event[transport.Msg]) error {
	if s.recvEventChan != nil {
		select {
		case s.recvEventChan <- event.Clone():
		default:
			logger.Errorf(s.gate.ctx, "session %q receive event channel is full", s.GetId())
		}
	}

	for i := range s.gate.options.SessionRecvEventHandlers {
		handler := s.gate.options.SessionRecvEventHandlers[i]
		if handler == nil {
			continue
		}
		err := internal.Call(func() error { return handler(s, event) })
		if err == nil || !errors.Is(err, protocol.ErrUnexpectedMsg) {
			return err
		}
	}

	for i := range s.recvEventHandlers {
		handler := s.recvEventHandlers[i]
		if handler == nil {
			continue
		}
		err := internal.Call(func() error { return handler(s, event) })
		if err == nil || !errors.Is(err, protocol.ErrUnexpectedMsg) {
			return err
		}
	}

	return protocol.ErrUnexpectedMsg
}

// PayloadHandler Payload消息事件处理器
func (s *_GtpSession) PayloadHandler(event protocol.Event[*transport.MsgPayload]) error {
	if s.recvDataChan != nil {
		select {
		case s.recvDataChan <- bytes.Clone(event.Msg.Data):
		default:
			logger.Errorf(s.gate.ctx, "session %q receive data channel is full", s.GetId())
		}
	}

	for i := range s.gate.options.SessionRecvDataHandlers {
		handler := s.gate.options.SessionRecvDataHandlers[i]
		if handler == nil {
			continue
		}
		err := internal.Call(func() error { return handler(s, event.Msg.Data) })
		if err == nil || !errors.Is(err, protocol.ErrUnexpectedMsg) {
			return err
		}
	}

	for i := range s.recvDataHandlers {
		handler := s.recvDataHandlers[i]
		if handler == nil {
			continue
		}
		err := internal.Call(func() error { return handler(s, event.Msg.Data) })
		if err == nil || !errors.Is(err, protocol.ErrUnexpectedMsg) {
			return err
		}
	}

	return protocol.ErrUnexpectedMsg
}
