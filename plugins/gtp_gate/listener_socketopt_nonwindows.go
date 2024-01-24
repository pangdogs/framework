//go:build !windows

package gtp_gate

import (
	"net"
	"syscall"
)

func newListenConfig(options *GateOptions) *net.ListenConfig {
	var noDelay *int
	if options.TCPNoDelay != nil {
		noDelay = types.New(types.Bool2Int[int](*options.TCPNoDelay))
	}

	var quickAck *int
	if options.TCPQuickAck != nil {
		quickAck = types.New(types.Bool2Int[int](*options.TCPQuickAck))
	}

	recvBuf := options.TCPRecvBuf
	sendBuf := options.TCPSendBuf
	lingerSec := options.TCPLinger

	return &net.ListenConfig{
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
