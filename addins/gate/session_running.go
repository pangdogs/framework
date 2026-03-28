/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package gate

import (
	"context"
	"errors"
	"fmt"
	"time"

	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/transport"
	"go.uber.org/zap"
)

// mainLoop 会话主线程
func (s *_Session) mainLoop() {
	defer func() {
		// 关闭连接和释放资源
		if s.transceiver.Conn != nil {
			s.transceiver.Conn.Close()
		}
		s.transceiver.Dispose()
		// 释放屏障
		s.gate.barrier.Done()
	}()

	// 调整会话状态为活跃
	s.setState(SessionState_Active)

	log.L(s.gate.svcCtx).Debug("session started", zap.String("session_id", s.Id().String()))

	pinged := false
	var timeout time.Time

	// 启动i/o发送线程
	go s.io.sendLoop()

loop:
	for {
		select {
		case <-s.Done():
			// 会话已关闭
			break loop

		case <-s.migrationChan:
			// 收到连接迁移通知
			addr := s.netAddr.Load()
			migrations := s.migrations.Add(1)

			err := transport.Retry{
				Transceiver: &s.transceiver,
				Times:       s.gate.options.IORetryTimes,
			}.Send(s.transceiver.Resend())

			log.L(s.gate.svcCtx).Debug("session connection migration",
				zap.String("session_id", s.Id().String()),
				zap.String("local", addr.Local.String()),
				zap.String("remote", addr.Remote.String()),
				zap.Int64("migrations", migrations),
				zap.NamedError("resend_error", err))

			// 调整会话状态为活跃，并重置ping状态
			s.setState(SessionState_Active)
			pinged = false

		default:
			// 长期处于非活跃状态时，检测超时并关闭会话
			if SessionState(s.state.Load()) == SessionState_Inactive {
				if time.Now().After(timeout) {
					s.close(&transport.RstError{
						Code:    gtp.Code_SessionDeath,
						Message: fmt.Sprintf("session death at %s", timeout.Format(time.RFC3339)),
					})
					continue
				}
			}
		}

		// 分发消息事件
		if err := s.eventDispatcher.Dispatch(s); err != nil {
			// 网络传输错误
			if errors.Is(err, transport.ErrTrans) {
				// 网络io错误
				if errors.Is(err, transport.ErrNetIO) {
					// 网络io超时，触发心跳检测，向对方发送ping
					if errors.Is(err, transport.ErrDeadlineExceeded) {
						if !pinged {
							// 尝试ping对端
							log.L(s.gate.svcCtx).Debug("session send ping", zap.String("session_id", s.Id().String()))
							s.ctrl.SendPing()
							pinged = true
						} else {
							// 未收到对方回复pong或其他消息事件，再次网络io超时，调整会话状态不活跃
							log.L(s.gate.svcCtx).Debug("session no pong received", zap.String("session_id", s.Id().String()))
							s.setState(SessionState_Inactive)
							timeout = time.Now().Add(s.gate.options.SessionInactiveTimeout)
						}
						continue
					}

					// 其他网络io类错误，调整会话状态不活跃，并重试
					log.L(s.gate.svcCtx).Error("session dispatching event failed, retry it",
						zap.String("session_id", s.Id().String()),
						zap.Error(err))
					s.setState(SessionState_Inactive)
					timeout = time.Now().Add(s.gate.options.SessionInactiveTimeout)
					continue
				}

				// 其他网络传输错误，关闭会话
				log.L(s.gate.svcCtx).Error("session dispatching event failed, close session",
					zap.String("session_id", s.Id().String()),
					zap.Error(err))
				s.close(&transport.RstError{
					Code:    gtp.Code_Reject,
					Message: err.Error(),
				})
				continue
			}

			// 非网络传输错误，丢弃不处理
			log.L(s.gate.svcCtx).Error("session dispatching event failed, discard it",
				zap.String("session_id", s.Id().String()),
				zap.Error(err))
		}

		// 没有错误，或非网络传输错误，调整会话状态为活跃，并重置ping状态
		s.setState(SessionState_Active)
		pinged = false
	}

	// 关闭会话
	s.close(nil)
	// 等待i/o线程结束
	<-s.io.terminated
	// 调整会话状态为已过期
	s.setState(SessionState_Death)
	// 发送关闭原因
	s.ctrl.SendRst(context.Cause(s))
	// 删除会话
	s.gate.deleteSession(s.Id())
	// 返回关闭结果
	async.ReturnVoid(s.closed)

	log.L(s.gate.svcCtx).Debug("session closed", zap.String("session_id", s.Id().String()))
}
