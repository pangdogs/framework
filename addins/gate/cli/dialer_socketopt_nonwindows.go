//go:build !windows

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
	"git.golaxy.org/core/utils/types"
	"net"
	"syscall"
)

func newDialer(options *ClientOptions) *net.Dialer {
	var noDelay *int
	if options.TCPNoDelay != nil {
		noDelay = types.NewCopiedT(types.Bool2Int[int](*options.TCPNoDelay))
	}

	var quickAck *int
	if options.TCPQuickAck != nil {
		quickAck = types.NewCopiedT(types.Bool2Int[int](*options.TCPQuickAck))
	}

	recvBuf := options.TCPRecvBuf
	sendBuf := options.TCPSendBuf
	lingerSec := options.TCPLinger

	return &net.Dialer{
		Control: func(network, address string, conn syscall.RawConn) error {
			return conn.Control(func(fd uintptr) {
				if noDelay != nil {
					syscall.SetsockoptInt(int(fd), syscall.SOL_TCP, syscall.TCP_NODELAY, *noDelay)
				}
				if quickAck != nil {
					syscall.SetsockoptInt(int(fd), syscall.SOL_TCP, syscall.TCP_QUICKACK, *quickAck)
				}
				if recvBuf != nil {
					syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_RCVBUF, *recvBuf)
				}
				if sendBuf != nil {
					syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_SNDBUF, *sendBuf)
				}
				if lingerSec != nil {
					var l syscall.Linger
					if *lingerSec >= 0 {
						l.Onoff = 1
						l.Linger = int32(*lingerSec)
					} else {
						l.Onoff = 0
						l.Linger = 0
					}
					syscall.SetsockoptLinger(int(fd), syscall.SOL_SOCKET, syscall.SO_LINGER, &l)
				}
			})
		},
		KeepAlive: -1,
	}
}
