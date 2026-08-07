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

// initConn 绑定首次握手建立的连接，并按协商序号初始化同步器和会话 ID。
func (c *Client) initConn(conn net.Conn, encoder *codec.Encoder, decoder *codec.Decoder, remoteSendSeq, remoteRecvSeq uint32, sessionId uid.Id) {
	// 采用服务端返回的对端序号初始化本地收发序号及补发缓存。
	c.transceiver.Conn = conn
	c.transceiver.Encoder = encoder
	c.transceiver.Decoder = decoder
	c.transceiver.Timeout = c.options.IOTimeout
	c.transceiver.Synchronizer = transport.NewSequencedSynchronizer(remoteRecvSeq, remoteSendSeq, c.options.IOBufferCap)

	// 地址快照与收发器同时提交，供并发查询。
	c.netAddr.Store(&NetAddr{
		Local:  conn.LocalAddr(),
		Remote: conn.RemoteAddr(),
	})

	// 首次握手完成后会话 ID 在客户端生命周期内保持不变。
	c.sessionId = sessionId
}

// migrateConn 串行提交新连接，更新地址与迁移计数，然后通知客户端主循环。
// 已有迁移正在进行时立即返回错误。
func (c *Client) migrateConn(conn net.Conn, remoteRecvSeq uint32) (sendSeq, recvSeq uint32, err error) {
	if !c.migrationMu.TryLock() {
		return 0, 0, errors.New("concurrent client connection migration rejected")
	}
	defer c.migrationMu.Unlock()

	// Transceiver 在收发锁内同步序号并原子切换底层连接。
	sendSeq, recvSeq, err = c.transceiver.Migrate(conn, remoteRecvSeq)
	if err != nil {
		return
	}

	// 迁移成功后立即更新计数，保证外部观察到的是已提交状态
	c.migrations.Add(1)

	// 迁移成功后发布新地址快照。
	c.netAddr.Store(&NetAddr{
		Local:  conn.LocalAddr(),
		Remote: conn.RemoteAddr(),
	})

	// 唤醒客户端主循环执行缓存补发并恢复活跃状态。
	select {
	case c.migrationChan <- struct{}{}:
	case <-c.Done():
		return 0, 0, c.Err()
	}

	return
}
