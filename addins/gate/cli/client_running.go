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

	"git.golaxy.org/framework/net/gtp/transport"
	"go.uber.org/zap"
)

// mainLoop 运行客户端的接收、心跳、重连、迁移和关闭主循环。
func (c *Client) mainLoop() {
	addr := c.NetAddr()
	c.logger.Debug("client started",
		zap.String("session_id", c.SessionID().String()),
		zap.String("user_id", c.UserID()),
		zap.String("endpoint", c.Endpoint()),
		zap.String("local", addr.Local.String()),
		zap.String("remote", addr.Remote.String()))

	active := true
	pinged := false
	var timeout time.Time

	// 启动 I/O 发送循环
	go c.io.sendLoop()

	autoReconnect := func() {
		c.logger.Debug("client auto reconnect started", zap.String("session_id", c.SessionID().String()))

		i := 0
		for ; c.options.AutoReconnectRetryTimes <= 0 || i < c.options.AutoReconnectRetryTimes; i++ {
			select {
			case <-c.Done():
				return
			default:
			}

			if err := Reconnect(c); err != nil {
				c.logger.Error("client auto reconnect failed",
					zap.String("session_id", c.SessionID().String()),
					zap.Int("retries", i+1),
					zap.Error(err))

				// 服务端返回 RST 或链路迁移失败时不再重试，直接关闭客户端
				var rstErr *transport.RstError
				if errors.As(err, &rstErr) || errors.Is(err, transport.ErrMigrateConn) {
					c.logger.Error("client auto reconnect aborted, close client", zap.String("session_id", c.SessionID().String()))
					c.close(err)
					return
				}

				// 重连失败，暂停一会再试
				time.Sleep(c.options.AutoReconnectInterval)
				continue
			}

			c.logger.Debug("client auto reconnect ok", zap.String("session_id", c.SessionID().String()), zap.Int("retries", i+1))
			return
		}

		c.logger.Error("client auto reconnect retries exhausted, close client", zap.String("session_id", c.SessionID().String()))
		c.close(ErrAutoReconnectRetriesExhausted)
	}

	changeActive := func(b bool) {
		old := active
		active = b
		if old != b && !b {
			if c.options.AutoReconnect {
				go autoReconnect()
			} else {
				timeout = time.Now().Add(c.options.InactiveTimeout)
			}
		}
	}

	handleMigration := func() {
		addr := c.netAddr.Load()
		migrations := c.migrations.Load()

		err := transport.Retry{
			Transceiver: &c.transceiver,
			Times:       c.options.IORetryTimes,
		}.Send(c.transceiver.Resend())

		c.logger.Debug("client connection migration",
			zap.String("session_id", c.SessionID().String()),
			zap.String("user_id", c.UserID()),
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
				// 网络 I/O 错误
				if errors.Is(err, transport.ErrNetIO) {
					// 网络 I/O 超时，触发心跳检测并向对端发送 Ping
					if errors.Is(err, transport.ErrDeadlineExceeded) {
						if !pinged {
							// 尝试 Ping 对端
							c.logger.Debug("client send ping", zap.String("session_id", c.SessionID().String()))
							c.ctrl.SendPing()
							pinged = true
						} else {
							// 未收到 Pong 或其他事件且再次 I/O 超时，将连接标记为不活跃
							c.logger.Debug("client no pong received", zap.String("session_id", c.SessionID().String()))
							changeActive(false)
						}
						continue
					}

					// 其他网络 I/O 错误将连接标记为不活跃，并按配置重连
					c.logger.Error("client dispatching event failed, retry it", zap.String("session_id", c.SessionID().String()))
					changeActive(false)
					continue
				}

				// 其他网络传输错误，关闭客户端
				c.logger.Error("client dispatching event failed, close client", zap.String("session_id", c.SessionID().String()))
				c.close(err)
				continue
			}

			// 非网络传输错误，丢弃不处理
			c.logger.Error("session dispatching event failed, discard it", zap.String("session_id", c.SessionID().String()))
		}

		// 没有错误或只有非传输错误时，将客户端标记为活跃并重置 Ping 状态
		changeActive(true)
		pinged = false
	}

	// 主循环退出后取消客户端，并等待异步发送队列排空。
	c.close(nil)
	c.correlation.Close()
	<-c.correlation.Done().Done()
	<-c.io.terminated.Signal().Done()

	// 发送队列停止后再释放连接和编解码资源。
	if c.transceiver.Conn != nil {
		c.transceiver.Conn.Close()
	}
	c.transceiver.Dispose()
	// 资源清理完成后兑现 Closed 信号。
	c.closed.Complete()

	c.logger.Debug("client closed", zap.String("session_id", c.SessionID().String()))
}
