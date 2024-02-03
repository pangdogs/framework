//go:build windows

package gate

import (
	"git.golaxy.org/core/util/types"
	"net"
	"syscall"
)

func newListenConfig(options *GateOptions) *net.ListenConfig {
	var noDelay *int
	if options.TCPNoDelay != nil {
		noDelay = types.New(types.Bool2Int[int](*options.TCPNoDelay))
	}

	recvBuf := options.TCPRecvBuf
	sendBuf := options.TCPSendBuf
	lingerSec := options.TCPLinger

	return &net.ListenConfig{
		Control: func(network, address string, conn syscall.RawConn) error {
			return conn.Control(func(fd uintptr) {
				if noDelay != nil {
					syscall.SetsockoptInt(syscall.Handle(fd), syscall.IPPROTO_TCP, syscall.TCP_NODELAY, *noDelay)
				}
				if recvBuf != nil {
					syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_RCVBUF, *recvBuf)
				}
				if sendBuf != nil {
					syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_SNDBUF, *sendBuf)
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
					syscall.SetsockoptLinger(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_LINGER, &l)
				}
			})
		},
		KeepAlive: -1,
	}
}
