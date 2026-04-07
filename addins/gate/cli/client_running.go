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

package cli

import (
	"errors"
	"time"

	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/framework/net/gtp/transport"
	"go.uber.org/zap"
)

// mainLoop 客户端主线程
func (c *Client) mainLoop() {
	addr := c.NetAddr()
	c.logger.Debug("client started",
		zap.String("session_id", c.SessionId().String()),
		zap.String("user_id", c.UserId()),
		zap.String("endpoint", c.Endpoint()),
		zap.String("local", addr.Local.String()),
		zap.String("remote", addr.Remote.String()))

	active := true
	pinged := false
	var timeout time.Time

	// 启动i/o线程
	go c.io.sendLoop()

	autoReconnect := func() {
		c.logger.Debug("client auto reconnect started", zap.String("session_id", c.SessionId().String()))

		i := 0
		for ; c.options.AutoReconnectRetryTimes <= 0 || i < c.options.AutoReconnectRetryTimes; i++ {
			select {
			case <-c.Done():
				return
			default:
			}

			if err := Reconnect(c); err != nil {
				c.logger.Error("client auto reconnect failed",
					zap.String("session_id", c.SessionId().String()),
					zap.Int("retries", i+1),
					zap.Error(err))

				// 服务端返回rst拒绝连接，恢复链路失败，这两种情况下不再重试，关闭客户端
				var rstErr *transport.RstError
				if errors.As(err, &rstErr) || errors.Is(err, transport.ErrMigrateConn) {
					c.logger.Error("client auto reconnect aborted, close client", zap.String("session_id", c.SessionId().String()))
					c.close(err)
					return
				}

				// 重连失败，暂停一会再试
				time.Sleep(c.options.AutoReconnectInterval)
				continue
			}

			c.logger.Debug("client auto reconnect ok", zap.String("session_id", c.SessionId().String()), zap.Int("retries", i+1))
			return
		}

		c.logger.Error("client auto reconnect retries exhausted, close client", zap.String("session_id", c.SessionId().String()))
		c.close(ErrAutoReconnectRetriesExhausted)
	}

	changeActive := func(b bool) {
		old := active
		active = b
		if old != b && !b {
			if c.options.AutoRecover {
				go autoReconnect()
			} else {
				timeout = time.Now().Add(c.options.InactiveTimeout)
			}
		}
	}

	handleMigration := func() {
		addr := c.netAddr.Load()
		migrations := c.migrations.Add(1)

		err := transport.Retry{
			Transceiver: &c.transceiver,
			Times:       c.options.IORetryTimes,
		}.Send(c.transceiver.Resend())

		c.logger.Debug("client connection migration",
			zap.String("session_id", c.SessionId().String()),
			zap.String("user_id", c.UserId()),
			zap.String("endpoint", c.Endpoint()),
			zap.String("local", addr.Local.String()),
			zap.String("remote", addr.Remote.String()),
			zap.Int64("migrations", migrations),
			zap.NamedError("resend_error", err))

		changeActive(true)
		pinged = false
	}

loop:
	for {
		// 长期处于非活跃状态时，并且未开启自动重连，检测超时并关闭客户端
		if !active {
			if c.options.AutoReconnect {
				select {
				case <-c.Done():
					break loop
				case <-c.migrationChan:
					handleMigration()
					continue
				}
			}

			wait := time.Until(timeout)
			if wait <= 0 {
				c.close(ErrInactiveTimeout)
				break loop
			}

			timer := time.NewTimer(wait)
			select {
			case <-c.Done():
				timer.Stop()
				break loop

			case <-c.migrationChan:
				timer.Stop()
				handleMigration()
				continue

			case <-timer.C:
				c.close(ErrInactiveTimeout)
				break loop
			}
		}

		select {
		case <-c.Done():
			break loop

		case <-c.migrationChan:
			handleMigration()
			continue

		default:
		}

		// 分发消息事件
		if err := c.eventDispatcher.Dispatch(c); err != nil {
			// 网络传输错误
			if errors.Is(err, transport.ErrTrans) {
				// 网络io错误
				if errors.Is(err, transport.ErrNetIO) {
					// 网络io超时，触发心跳检测，向对方发送ping
					if errors.Is(err, transport.ErrDeadlineExceeded) {
						if !pinged {
							// 尝试ping对端
							c.logger.Debug("client send ping", zap.String("session_id", c.SessionId().String()))
							c.ctrl.SendPing()
							pinged = true
						} else {
							// 未收到对方回复pong或其他消息事件，再次网络io超时，调整连接状态不活跃
							c.logger.Debug("client no pong received", zap.String("session_id", c.SessionId().String()))
							changeActive(false)
						}
						continue
					}

					// 其他网络io类错误，调整连接状态不活跃，并重试
					c.logger.Error("client dispatching event failed, retry it", zap.String("session_id", c.SessionId().String()))
					changeActive(false)
					continue
				}

				// 其他网络传输错误，关闭客户端
				c.logger.Error("client dispatching event failed, close client", zap.String("session_id", c.SessionId().String()))
				c.close(err)
				continue
			}

			// 非网络传输错误，丢弃不处理
			c.logger.Error("session dispatching event failed, discard it", zap.String("session_id", c.SessionId().String()))
		}

		// 没有错误，或非网络传输错误，调整客户端状态为活跃，并重置ping状态
		changeActive(true)
		pinged = false
	}

	// 关闭客户端
	c.close(nil)
	// 等待i/o线程结束
	<-c.io.terminated
	// 关闭网络连接
	if c.transceiver.Conn != nil {
		c.transceiver.Conn.Close()
	}
	c.transceiver.Dispose()
	// 返回关闭结果
	async.ReturnVoid(c.closed)

	c.logger.Debug("client closed", zap.String("session_id", c.SessionId().String()))
}
