package gtp_gate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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
	s.transceiver.Synchronizer = transport.NewSequencedSynchronizer(rand.Uint32(), rand.Uint32(), s.gate.options.IOBufferCap)

	// 初始化刷新通知channel
	s.renewChan = make(chan struct{}, 1)

	// 初始化token
	s.token = token

	return s.transceiver.Synchronizer.SendSeq(), s.transceiver.Synchronizer.RecvSeq()
}

// renew 刷新
func (s *_Session) renew(conn net.Conn, remoteRecvSeq uint32) (sendSeq, recvSeq uint32, err error) {
	s.Lock()
	defer s.Unlock()

	// 刷新链路
	sendSeq, recvSeq, err = s.transceiver.Renew(conn, remoteRecvSeq)
	if err != nil {
		return
	}

	// 通知刷新
	select {
	case s.renewChan <- struct{}{}:
	default:
	}

	return
}

// pauseIO 暂停收发消息
func (s *_Session) pauseIO() {
	s.transceiver.Pause()
}

// continueIO 继续收发消息
func (s *_Session) continueIO() {
	s.transceiver.Continue()
}

// mainLoop 主线程
func (s *_Session) mainLoop() {
	defer func() {
		s.cancel(nil)

		// 调整会话状态为已过期
		s.setState(SessionState_Death)

		// 关闭连接和清理数据
		if s.transceiver.Conn != nil {
			s.transceiver.Conn.Close()
		}
		s.transceiver.Clean()

		// 删除会话
		s.gate.deleteSession(s.GetId())

		s.gate.wg.Done()
		close(s.closedChan)
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

loop:
	for {
		// 非活跃状态，检测超时时间
		if s.state == SessionState_Inactive {
			if time.Now().After(timeout) {
				s.cancel(&transport.RstError{
					Code:    gtp.Code_SessionDeath,
					Message: fmt.Sprintf("session death at %s", timeout.Format(time.RFC3339)),
				})
				break loop
			}
		}

		// 检测会话是否已关闭
		select {
		case <-s.Done():
			break loop
		case <-s.gate.ctx.Done():
			s.cancel(&transport.RstError{
				Code:    gtp.Code_Shutdown,
				Message: "service shutdown",
			})
			break loop
		default:
		}

		// 分发消息事件
		if err := s.eventDispatcher.Dispatching(); err != nil {
			// 网络io错误
			if errors.Is(err, transport.ErrNetIO) {
				// 网络io超时，触发心跳检测，向对方发送ping
				if errors.Is(err, transport.ErrTimeout) {
					if !pinged {
						// 尝试ping对端
						log.Debugf(s.gate.ctx, "session %q send ping", s.GetId())
						s.ctrl.SendPing()
						pinged = true
					} else {
						// 未收到对方回复pong或其他消息事件，再次网络io超时，调整会话状态不活跃
						log.Debugf(s.gate.ctx, "session %q no pong received", s.GetId())
						if s.setState(SessionState_Inactive) {
							timeout = time.Now().Add(s.gate.options.SessionInactiveTimeout)
						}
					}
					continue
				}

				// 其他网络io类错误，调整会话状态不活跃
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

			log.Errorf(s.gate.ctx, "session %q dispatching event failed, %s", s.GetId(), err)
			continue
		}

		// 没有错误，或非网络io类错误，重置ping状态
		pinged = false
		// 调整会话状态活跃
		s.setState(SessionState_Active)
	}

	// 发送关闭原因
	s.ctrl.SendRst(context.Cause(s))

	log.Debugf(s.gate.ctx, "session %q shutdown, conn %q -> %q", s.GetId(), s.GetLocalAddr(), s.GetRemoteAddr())
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
			log.Errorf(s.gate.ctx, "handle session %q state changed failed, %s", s.GetId(), panicErr)
		}
		return false
	}

	s.options.StateChangedHandler.Invoke(interrupt, old, state)
	s.gate.options.SessionStateChangedHandler.Invoke(interrupt, s, old, state)

	return true
}

// handleEvent 接收自定义事件的处理器
func (s *_Session) handleEvent(event transport.Event[gtp.Msg]) error {
	// 写入channel
	if s.options.RecvEventChan != nil {
		eventCopy := event
		eventCopy.Msg = eventCopy.Msg.Clone()

		select {
		case s.options.RecvEventChan <- eventCopy:
		default:
			log.Errorf(s.gate.ctx, "session %q receive event channel is full", s.GetId())
		}
	}

	var errs []error

	interrupt := func(err, _ error) bool {
		if err != nil {
			errs = append(errs, err)
		}
		return false
	}

	// 回调监控器
	s.eventWatchers.AutoRLock(func(watchers *[]*_EventWatcher) {
		for i := range *watchers {
			(*watchers)[i].handler.Exec(interrupt, event)
		}
	})

	// 回调会话处理器
	s.options.RecvEventHandler.Exec(interrupt, event)

	// 回调网关处理器
	s.gate.options.SessionRecvEventHandler.Exec(interrupt, s, event)

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// handlePayload Payload消息事件处理器
func (s *_Session) handlePayload(event transport.Event[gtp.MsgPayload]) error {
	// 写入channel
	if s.options.RecvDataChan != nil {
		select {
		case s.options.RecvDataChan <- bytes.Clone(event.Msg.Data):
		default:
			log.Errorf(s.gate.ctx, "session %q receive data channel is full", s.GetId())
		}
	}

	var errs []error

	interrupt := func(err, _ error) bool {
		if err != nil {
			errs = append(errs, err)
		}
		return false
	}

	// 回调监控器
	s.dataWatchers.AutoRLock(func(watchers *[]*_DataWatcher) {
		for i := range *watchers {
			(*watchers)[i].handler.Exec(interrupt, event.Msg.Data)
		}
	})

	// 回调会话处理器
	s.options.RecvDataHandler.Invoke(interrupt, event.Msg.Data)

	// 回调网关处理器
	s.gate.options.SessionRecvDataHandler.Invoke(interrupt, s, event.Msg.Data)

	if len(errs) > 0 {
		return errors.Join(errs...)
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
