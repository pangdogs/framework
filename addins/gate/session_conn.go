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
	"errors"
	"math/rand"
	"net"

	"git.golaxy.org/framework/net/gtp/codec"
	"git.golaxy.org/framework/net/gtp/transport"
)

// initConn 绑定首次握手建立的连接，以随机序号初始化同步器并返回双方起始序号。
func (s *_Session) initConn(conn net.Conn, encoder *codec.Encoder, decoder *codec.Decoder) (sendSeq, recvSeq uint32) {
	// 首次连接使用随机收发序号，并保留未确认帧用于后续迁移补发。
	s.transceiver.Conn = conn
	s.transceiver.Encoder = encoder
	s.transceiver.Decoder = decoder
	s.transceiver.Timeout = s.gate.options.IOTimeout
	s.transceiver.Synchronizer = transport.NewSequencedSynchronizer(rand.Uint32(), rand.Uint32(), s.gate.options.IOBufferCap)

	// 地址快照与收发器同时提交，供并发查询。
	s.netAddr.Store(&NetAddr{
		Local:  conn.LocalAddr(),
		Remote: conn.RemoteAddr(),
	})

	return s.transceiver.Synchronizer.SendSeq(), s.transceiver.Synchronizer.RecvSeq()
}

// migrateConn 串行提交新连接，更新地址与迁移计数，然后通知会话主循环。
// 已有迁移正在进行时立即返回错误。
func (s *_Session) migrateConn(conn net.Conn, remoteRecvSeq uint32) (sendSeq, recvSeq uint32, err error) {
	if !s.migrationMu.TryLock() {
		return 0, 0, errors.New("concurrent session connection migration rejected")
	}
	defer s.migrationMu.Unlock()

	// Transceiver 在收发锁内同步序号并原子切换底层连接。
	sendSeq, recvSeq, err = s.transceiver.Migrate(conn, remoteRecvSeq)
	if err != nil {
		return
	}

	// 迁移成功后立即更新计数，保证外部观察到的是已提交状态
	s.migrations.Add(1)

	// 迁移成功后发布新地址快照。
	s.netAddr.Store(&NetAddr{
		Local:  conn.LocalAddr(),
		Remote: conn.RemoteAddr(),
	})

	// 唤醒会话主循环执行缓存补发并恢复活跃状态。
	select {
	case s.migrationChan <- struct{}{}:
	case <-s.Done():
		return 0, 0, s.Err()
	}

	return
}
