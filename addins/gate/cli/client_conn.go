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
	"net"

	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/net/gtp/codec"
	"git.golaxy.org/framework/net/gtp/transport"
)

// initConn 初始化连接
func (c *Client) initConn(conn net.Conn, encoder *codec.Encoder, decoder *codec.Decoder, remoteSendSeq, remoteRecvSeq uint32, sessionId uid.Id) {
	// 初始化消息收发器
	c.transceiver.Conn = conn
	c.transceiver.Encoder = encoder
	c.transceiver.Decoder = decoder
	c.transceiver.Timeout = c.options.IOTimeout
	c.transceiver.Synchronizer = transport.NewSequencedSynchronizer(remoteRecvSeq, remoteSendSeq, c.options.IOBufferCap)

	// 记录连接信息
	c.netAddr.Store(&NetAddr{
		Local:  conn.LocalAddr(),
		Remote: conn.RemoteAddr(),
	})

	// 记录会话Id
	c.sessionId = sessionId
}

// migrateConn 迁移连接
func (c *Client) migrateConn(conn net.Conn, remoteRecvSeq uint32) (sendSeq, recvSeq uint32, err error) {
	if !c.migrationMutex.TryLock() {
		return 0, 0, errors.New("concurrent client connection migration rejected")
	}
	defer c.migrationMutex.Unlock()

	// 迁移连接
	sendSeq, recvSeq, err = c.transceiver.Migrate(conn, remoteRecvSeq)
	if err != nil {
		return
	}

	// 记录连接信息
	c.netAddr.Store(&NetAddr{
		Local:  conn.LocalAddr(),
		Remote: conn.RemoteAddr(),
	})

	// 通知连接已迁移
	select {
	case c.migrationChan <- struct{}{}:
	case <-c.Done():
		return 0, 0, c.Err()
	}

	return
}
