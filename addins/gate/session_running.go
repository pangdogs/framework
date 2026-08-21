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

	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/transport"
	"go.uber.org/zap"
)

// mainLoop 运行会话的接收、心跳、迁移和关闭主循环。
func (s *_Session) mainLoop() {
	defer s.gate.barrier.Done()

	// 调整会话状态为活跃
	s.setState(SessionState_Active)

	log.L(s.gate.svcCtx).Debug("session started", zap.String("session_id", s.ID().String()))

	pinged := false
	var timeout time.Time

	// 启动 I/O 发送循环
	go s.io.sendLoop()

	handleMigration := func() {
		addr := s.netAddr.Load()
		migrations := s.migrations.Load()

		err := transport.Retry{
			Transceiver: &s.transceiver,
			Times:       s.gate.options.IORetryTimes,
		}.Send(s.transceiver.Resend())

		log.L(s.gate.svcCtx).Debug("session connection migration",
			zap.String("session_id", s.ID().String()),
			zap.String("local", addr.Local.String()),
			zap.String("remote", addr.Remote.String()),
			zap.Int64("migrations", migrations),
			zap.NamedError("resend_error", err))

		// 调整会话状态为活跃，并重置 Ping 状态
		s.setState(SessionState_Active)
		pinged = false
	}

	closeInactive := func() {
		s.asyncScope.Close(&transport.RstError{
			Code:    gtp.Code_SessionDeath,
			Message: fmt.Sprintf("session death at %s", timeout.Format(time.RFC3339)),
		})
	}

loop:
	for {
		// 长期处于非活跃状态时，检测超时并关闭会话
		if SessionState(s.state.Load()) == SessionState_Inactive {
			wait := time.Until(timeout)
			if wait <= 0 {
				closeInactive()
				break loop
			}

			timer := time.NewTimer(wait)
			select {
			case <-s.Done():
				timer.Stop()
				break loop

			case <-s.migrationChan:
				timer.Stop()
				handleMigration()
				continue

			case <-timer.C:
				closeInactive()
				break loop
			}
		}

		select {
		case <-s.Done():
			break loop

		case <-s.migrationChan:
			handleMigration()
			continue

		default:
		}

		// 分发消息事件
		if err := s.eventDispatcher.Dispatch(s); err != nil {
			// 网络传输错误
			if errors.Is(err, transport.ErrTrans) {
				// 网络 I/O 错误
				if errors.Is(err, transport.ErrNetIO) {
					// 网络 I/O 超时，触发心跳检测并向对端发送 Ping
					if errors.Is(err, transport.ErrDeadlineExceeded) {
						if !pinged {
							// 尝试 Ping 对端
							log.L(s.gate.svcCtx).Debug("session send ping", zap.String("session_id", s.ID().String()))
							s.ctrl.SendPing()
							pinged = true
						} else {
							// 未收到 Pong 或其他事件且再次 I/O 超时，将会话标记为不活跃
							log.L(s.gate.svcCtx).Debug("session no pong received", zap.String("session_id", s.ID().String()))
							s.setState(SessionState_Inactive)
							timeout = time.Now().Add(s.gate.options.SessionInactiveTimeout)
						}
						continue
					}

					// 其他网络 I/O 错误将会话标记为不活跃，等待连接迁移
					log.L(s.gate.svcCtx).Error("session dispatching event failed, retry it",
						zap.String("session_id", s.ID().String()),
						zap.Error(err))
					s.setState(SessionState_Inactive)
					timeout = time.Now().Add(s.gate.options.SessionInactiveTimeout)
					continue
				}

				// 其他网络传输错误，关闭会话
				log.L(s.gate.svcCtx).Error("session dispatching event failed, close session",
					zap.String("session_id", s.ID().String()),
					zap.Error(err))
				s.asyncScope.Close(&transport.RstError{
					Code:    gtp.Code_Reject,
					Message: err.Error(),
				})
				continue
			}

			// 非网络传输错误，丢弃不处理
			log.L(s.gate.svcCtx).Error("session dispatching event failed, discard it",
				zap.String("session_id", s.ID().String()),
				zap.Error(err))
		}

		// 没有错误或只有非传输错误时，将会话标记为活跃并重置 Ping 状态
		s.setState(SessionState_Active)
		pinged = false
	}

	// 主循环退出后关闭会话异步作用域，并等待异步发送队列排空。
	s.asyncScope.Close()
	<-s.io.terminated.Signal().Done()
	s.setState(SessionState_Death)

	// 尽力向对端发送关闭原因，再撤销注册并释放连接资源。
	s.ctrl.SendRst(context.Cause(s))
	s.gate.deleteSession(s.ID())
	if s.transceiver.Conn != nil {
		s.transceiver.Conn.Close()
	}
	s.transceiver.Dispose()
	// 资源清理完成后兑现 Closed 信号。
	s.closed.Complete()

	log.L(s.gate.svcCtx).Debug("session closed", zap.String("session_id", s.ID().String()))
}
