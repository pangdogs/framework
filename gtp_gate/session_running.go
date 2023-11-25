package gtp_gate

import (
	"bytes"
	"errors"
	"fmt"
	"kit.golaxy.org/golaxy"
	"kit.golaxy.org/golaxy/util/generic"
	"kit.golaxy.org/golaxy/util/types"
	"kit.golaxy.org/plugins/gtp"
	"kit.golaxy.org/plugins/gtp/codec"
	"kit.golaxy.org/plugins/gtp/transport"
	"kit.golaxy.org/plugins/log"
	"math/rand"
	"net"
	"time"
)

// init 初始化
func (s *_Session) init(conn net.Conn, encoder codec.IEncoder, decoder codec.IDecoder, token string) (sendSeq, recvSeq uint32) {
	s.Lock()
	defer s.Unlock()

	// 初始化消息收发器
	s.transceiver.Conn = conn
	s.transceiver.Encoder = encoder
	s.transceiver.Decoder = decoder
	s.transceiver.Timeout = s.gate.options.IOTimeout

	buff := &transport.SequencedBuffer{}
	buff.Reset(rand.Uint32(), rand.Uint32(), s.gate.options.IOBufferCap)

	s.transceiver.Buffer = buff

	// 初始化刷新通知channel
	s.renewChan = make(chan struct{}, 1)

	// 初始化token
	s.token = token

	return buff.SendSeq(), buff.RecvSeq()
}

// renew 刷新
func (s *_Session) renew(conn net.Conn, remoteRecvSeq uint32) (sendSeq, recvSeq uint32, err error) {
	s.Lock()

	defer func() {
		s.Unlock()

		select {
		case s.renewChan <- struct{}{}:
		default:
		}
	}()

	// 刷新链路
	return s.transceiver.Renew(conn, remoteRecvSeq)
}

// pauseIO 暂停收发消息
func (s *_Session) pauseIO() {
	s.transceiver.Pause()
}

// continueIO 继续收发消息
func (s *_Session) continueIO() {
	s.transceiver.Continue()
}

// run 运行（会话的主线程）
func (s *_Session) run() {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			defer s.cancel()
			log.Errorf(s.gate.ctx, "session %q panicked, %s", s.GetId(), fmt.Errorf("%w: %w", golaxy.ErrPanicked, panicErr))
		}

		// 调整会话状态为已过期
		s.setState(SessionState_Death)

		// 关闭连接和清理数据
		if s.transceiver.Conn != nil {
			s.transceiver.Conn.Close()
		}
		s.transceiver.Clean()

		// 删除会话
		s.gate.deleteSession(s.GetId())

		log.Debugf(s.gate.ctx, "session %q shutdown, conn %q -> %q", s.GetId(), s.GetLocalAddr(), s.GetRemoteAddr())
	}()

	log.Debugf(s.gate.ctx, "session %q started, conn %q -> %q", s.GetId(), s.GetLocalAddr(), s.GetRemoteAddr())

	// 启动发送数据的线程
	if s.options.SendDataChan != nil {
		go func() {
			for {
				select {
				case data := <-s.options.SendDataChan:
					if err := s.SendData(data); err != nil {
						log.Errorf(s.gate.ctx, "session %q fetch data from the send data channel for sending failed, %s", s.GetId(), err)
					}
				case <-s.Done():
					return
				}
			}
		}()
	}

	// 启动发送自定义事件的线程
	if s.options.SendEventChan != nil {
		go func() {
			for {
				select {
				case event := <-s.options.SendEventChan:
					if err := s.SendEvent(event); err != nil {
						log.Errorf(s.gate.ctx, "session %q fetch event from the send event channel for sending failed, %s", s.GetId(), err)
					}
				case <-s.Done():
					return
				}
			}
		}()
	}

	pinged := false
	var timeout time.Time

	// 调整会话状态为活跃
	s.setState(SessionState_Active)

	for {
		// 非活跃状态，检测超时时间
		if s.state == SessionState_Inactive {
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
		if err := s.eventDispatcher.Dispatching(); err != nil {
			log.Debugf(s.gate.ctx, "session %q dispatching event failed, %s", s.GetId(), err)

			// 网络io超时，触发心跳检测，向对方发送ping
			if errors.Is(err, transport.ErrTimeout) {
				if !pinged {
					log.Debugf(s.gate.ctx, "session %q send ping", s.GetId())

					s.ctrl.SendPing()
					pinged = true
				} else {
					log.Debugf(s.gate.ctx, "session %q no pong received", s.GetId())

					// 未收到对方回复pong或其他消息事件，再次网络io超时，调整会话状态不活跃
					if s.setState(SessionState_Inactive) {
						timeout = time.Now().Add(s.gate.options.SessionInactiveTimeout)
					}
				}
				continue
			}

			// 其他网络io类错误，调整会话状态不活跃
			if errors.Is(err, transport.ErrNetIO) {
				if s.setState(SessionState_Inactive) {
					timeout = time.Now().Add(s.gate.options.SessionInactiveTimeout)
				}

				func() {
					timer := time.NewTimer(10 * time.Second)
					defer timer.Stop()

					select {
					case <-timer.C:
						return
					case <-s.renewChan:
						// 发送缓存的消息
						transport.Retry{
							Transceiver: &s.transceiver,
							Times:       s.gate.options.IORetryTimes,
						}.Send(s.transceiver.Resend())
						// 重置ping状态
						pinged = false
						return
					case <-s.Done():
						return
					}
				}()

				log.Debugf(s.gate.ctx, "session %q retry dispatching event, conn %q -> %q", s.GetId(), s.GetLocalAddr(), s.GetRemoteAddr())
				continue
			}

			continue
		}

		// 没有错误，或非网络io类错误，重置ping状态
		pinged = false
		// 调整会话状态活跃
		s.setState(SessionState_Active)
	}
}

// setState 调整会话状态
func (s *_Session) setState(state SessionState) bool {
	old := s.state

	if old == state {
		return false
	}

	s.Lock()
	s.state = state
	s.Unlock()

	log.Debugf(s.gate.ctx, "session %q state %q => %q", s.GetId(), old, state)

	interrupt := func(panicErr error) bool {
		if panicErr != nil {
			log.Errorf(s.gate.ctx, "session %q state changed handler error: %s", s.GetId(), panicErr)
		}
		return false
	}

	s.gate.options.SessionStateChangedHandler.Invoke(interrupt, s, old, state)
	s.options.StateChangedHandler.Invoke(interrupt, old, state)

	return true
}

// handleEvent 接收自定义事件的处理器
func (s *_Session) handleEvent(event transport.Event[gtp.Msg]) error {
	if s.options.RecvEventChan != nil {
		eventCopy := event
		eventCopy.Msg = eventCopy.Msg.Clone()

		select {
		case s.options.RecvEventChan <- eventCopy:
		default:
			log.Errorf(s.gate.ctx, "session %q receive event channel is full", s.GetId())
		}
	}

	interrupt := func(err, panicErr error) bool {
		err = generic.FuncError(err, panicErr)
		if err == nil || !errors.Is(err, transport.ErrUnexpectedMsg) {
			if err != nil {
				log.Errorf(s.gate.ctx, "session %q receive event handler error: %s", s.GetId(), err)
			}
			return true
		}
		return false
	}

	err1 := generic.FuncError(s.gate.options.SessionRecvEventHandler.Invoke(interrupt, s, event))
	err2 := generic.FuncError(s.options.RecvEventHandler.Invoke(interrupt, event))

	if errors.Is(err1, transport.ErrUnexpectedMsg) && errors.Is(err2, transport.ErrUnexpectedMsg) {
		return transport.ErrUnexpectedMsg
	}

	return nil
}

// handlePayload Payload消息事件处理器
func (s *_Session) handlePayload(event transport.Event[gtp.MsgPayload]) error {
	if s.options.RecvDataChan != nil {
		select {
		case s.options.RecvDataChan <- bytes.Clone(event.Msg.Data):
		default:
			log.Errorf(s.gate.ctx, "session %q receive data channel is full", s.GetId())
		}
	}

	interrupt := func(err, panicErr error) bool {
		err = generic.FuncError(err, panicErr)
		if err == nil || !errors.Is(err, transport.ErrUnexpectedMsg) {
			if err != nil {
				log.Errorf(s.gate.ctx, "session %q receive data handler error: %s", s.GetId(), err)
			}
			return true
		}
		return false
	}

	err1 := generic.FuncError(s.gate.options.SessionRecvDataHandler.Invoke(interrupt, s, event.Msg.Data))
	err2 := generic.FuncError(s.options.RecvDataHandler.Invoke(interrupt, event.Msg.Data))

	if errors.Is(err1, transport.ErrUnexpectedMsg) && errors.Is(err2, transport.ErrUnexpectedMsg) {
		return transport.ErrUnexpectedMsg
	}

	return nil
}

// handleHeartbeat Heartbeat消息事件处理器
func (s *_Session) handleHeartbeat(event transport.Event[gtp.MsgHeartbeat]) error {
	if event.Flags.Is(gtp.Flag_Ping) {
		log.Debugf(s.gate.ctx, "session %q receive ping", s.GetId())
	} else {
		log.Debugf(s.gate.ctx, "session %q receive pong", s.GetId())
	}
	return nil
}
