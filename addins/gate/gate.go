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
	"context"
	"crypto/tls"
	"errors"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/transport"
	"git.golaxy.org/framework/utils/concurrent"
	"golang.org/x/net/websocket"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
)

// IWatcher 监听器
type IWatcher interface {
	context.Context
	Terminate() <-chan struct{}
	Terminated() <-chan struct{}
}

// IGate 网关
type IGate interface {
	// GetSession 查询会话
	GetSession(sessionId uid.Id) (ISession, bool)
	// RangeSessions 遍历所有会话
	RangeSessions(fun generic.Func1[ISession, bool])
	// EachSessions 遍历所有会话
	EachSessions(fun generic.Action1[ISession])
	// CountSessions 统计所有会话数量
	CountSessions() int
	// Watch 监听会话变化
	Watch(ctx context.Context, handler SessionStateChangedHandler) IWatcher
}

func newGate(settings ...option.Setting[GateOptions]) IGate {
	return &_Gate{
		options: option.Make(With.Default(), settings...),
	}
}

type _Gate struct {
	svcCtx          service.Context
	ctx             context.Context
	terminate       context.CancelCauseFunc
	options         GateOptions
	wg              sync.WaitGroup
	tcpListener     net.Listener
	wsListener      *http.Server
	sessionMap      sync.Map
	sessionCount    int64
	sessionWatchers concurrent.LockedSlice[*_SessionWatcher]
}

// Init 初始化插件
func (g *_Gate) Init(svcCtx service.Context, _ runtime.Context) {
	log.Infof(svcCtx, "init addin %q", self.Name)

	g.svcCtx = svcCtx
	g.ctx, g.terminate = context.WithCancelCause(context.Background())

	if g.options.TCPAddress != "" {
		listener, err := newListenConfig(&g.options).Listen(context.Background(), "tcp", g.options.TCPAddress)
		if err != nil {
			log.Panicf(g.svcCtx, "listen tcp %q failed, %s", g.options.TCPAddress, err)
		}

		if g.options.TCPTLSConfig != nil {
			listener = tls.NewListener(listener, g.options.TCPTLSConfig)
		}

		g.tcpListener = listener

		log.Infof(g.svcCtx, "listener %q started", g.tcpListener.Addr())

		go func() {
			for {
				conn, err := g.tcpListener.Accept()
				if err != nil {
					if errors.Is(err, net.ErrClosed) {
						log.Debugf(g.svcCtx, "listener %q closed", g.tcpListener.Addr())
						return
					}
					log.Errorf(g.svcCtx, "listener %q accept a new connection failed, %s", g.tcpListener.Addr(), err)
					continue
				}

				log.Debugf(g.svcCtx, "listener %q accept a new connection, remote %q", g.tcpListener.Addr(), conn.RemoteAddr())
				go g.handleSession(conn)
			}
		}()
	}

	if g.options.WebSocketURL != nil {
		listener := &http.Server{
			Addr:         g.options.WebSocketURL.Host,
			ReadTimeout:  g.options.IOTimeout,
			WriteTimeout: g.options.IOTimeout,
			IdleTimeout:  g.options.IOTimeout,
		}

		if strings.EqualFold(g.options.WebSocketURL.Scheme, "https") || strings.EqualFold(g.options.WebSocketURL.Scheme, "wss") {
			listener.TLSConfig = g.options.WebSocketTLSConfig
			if listener.TLSConfig == nil {
				log.Panicf(g.svcCtx, "use HTTPS to listen, need to provide a valid TLS configuration")
			}
		}

		mux := http.NewServeMux()
		mux.Handle(g.options.WebSocketURL.Path, websocket.Handler(func(conn *websocket.Conn) {
			log.Debugf(g.svcCtx, "listener %q accept a new connection, remote %q", conn.LocalAddr(), conn.RemoteAddr())
			if session, ok := g.handleSession(conn); ok {
				<-session.Closed()
			}
		}))
		listener.Handler = mux

		g.wsListener = listener

		log.Infof(g.svcCtx, "listener %q started", g.options.WebSocketURL)

		go func() {
			if err := g.wsListener.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Panicf(g.svcCtx, "listener %q was interrupted, %s", g.options.WebSocketURL, err)
			}
		}()
	}

	if g.tcpListener == nil && g.wsListener == nil {
		log.Panic(g.svcCtx, "no address need to listen")
	}
}

// Shut 关闭插件
func (g *_Gate) Shut(svcCtx service.Context, _ runtime.Context) {
	log.Infof(svcCtx, "shut addin %q", self.Name)

	g.terminate(&transport.RstError{
		Code:    gtp.Code_Shutdown,
		Message: "service shutdown",
	})

	g.wg.Wait()

	if g.tcpListener != nil {
		g.tcpListener.Close()
	}
	if g.wsListener != nil {
		g.wsListener.Close()
	}
}

// GetSession 查询会话
func (g *_Gate) GetSession(sessionId uid.Id) (ISession, bool) {
	return g.getSession(sessionId)
}

// RangeSessions 遍历所有会话
func (g *_Gate) RangeSessions(fun generic.Func1[ISession, bool]) {
	g.sessionMap.Range(func(k, v any) bool {
		return fun.Exec(v.(ISession))
	})
}

// EachSessions 遍历所有会话
func (g *_Gate) EachSessions(fun generic.Action1[ISession]) {
	g.sessionMap.Range(func(k, v any) bool {
		fun.Exec(v.(ISession))
		return true
	})
}

// CountSessions 统计所有会话数量
func (g *_Gate) CountSessions() int {
	return int(atomic.LoadInt64(&g.sessionCount))
}

// Watch 监听会话变化
func (g *_Gate) Watch(ctx context.Context, handler SessionStateChangedHandler) IWatcher {
	return g.newSessionWatcher(ctx, handler)
}
