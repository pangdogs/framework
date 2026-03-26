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

// initConn 初始化连接
func (s *_Session) initConn(conn net.Conn, encoder *codec.Encoder, decoder *codec.Decoder) (sendSeq, recvSeq uint32) {
	// 初始化消息收发器
	s.transceiver.Conn = conn
	s.transceiver.Encoder = encoder
	s.transceiver.Decoder = decoder
	s.transceiver.Timeout = s.gate.options.IOTimeout
	s.transceiver.Synchronizer = transport.NewSequencedSynchronizer(rand.Uint32(), rand.Uint32(), s.gate.options.IOBufferCap)

	// 记录连接信息
	s.netAddr.Store(&NetAddr{
		Local:  conn.LocalAddr(),
		Remote: conn.RemoteAddr(),
	})

	return s.transceiver.Synchronizer.SendSeq(), s.transceiver.Synchronizer.RecvSeq()
}

// migrateConn 迁移连接
func (s *_Session) migrateConn(conn net.Conn, remoteRecvSeq uint32) (sendSeq, recvSeq uint32, err error) {
	if !s.migrationMutex.TryLock() {
		return 0, 0, errors.New("concurrent session connection migration rejected")
	}
	defer s.migrationMutex.Unlock()

	// 迁移连接
	sendSeq, recvSeq, err = s.transceiver.Migrate(conn, remoteRecvSeq)
	if err != nil {
		return
	}

	// 记录连接信息
	s.netAddr.Store(&NetAddr{
		Local:  conn.LocalAddr(),
		Remote: conn.RemoteAddr(),
	})

	// 通知连接已迁移
	select {
	case s.migrationChan <- struct{}{}:
	case <-s.Done():
		return 0, 0, s.Err()
	}

	return
}
