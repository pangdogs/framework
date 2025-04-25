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
	"golang.org/x/net/websocket"
	"net"
	"strings"
)

func DefaultWebSocketLocalAddrResolver(conn *websocket.Conn) net.Addr {
	return conn.LocalAddr()
}

func DefaultWebSocketRemoteAddrResolver(conn *websocket.Conn) net.Addr {
	if xrip := conn.Request().Header.Get("X-Real-Ip"); xrip != "" {
		return _WebSocketAddr(strings.TrimSpace(xrip))
	}

	if xff := conn.Request().Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return _WebSocketAddr(strings.TrimSpace(ips[0]))
		}
	}

	ip, _, err := net.SplitHostPort(conn.Request().RemoteAddr)
	if err == nil {
		return _WebSocketAddr(ip)
	}

	return _WebSocketAddr("unknown")
}

type _WebSocketAddr string

func (_WebSocketAddr) Network() string {
	return "websocket"
}

func (addr _WebSocketAddr) String() string {
	return string(addr)
}

type _WebSocketConn struct {
	*websocket.Conn
	gate *_Gate
}

func (c *_WebSocketConn) LocalAddr() net.Addr {
	return c.gate.options.WebSocketLocalAddrResolver(c.Conn)
}

func (c *_WebSocketConn) RemoteAddr() net.Addr {
	return c.gate.options.WebSocketRemoteAddrResolver(c.Conn)
}
