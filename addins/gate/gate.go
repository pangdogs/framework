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
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/transport"
	"git.golaxy.org/framework/utils/concurrent"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
)

// IGate 提供已鉴权客户端会话的并发查询与建立监听能力。
type IGate interface {
	// Get 按会话 ID 查询当前尚未过期的会话。
	Get(id uid.Id) (ISession, bool)
	// Count 返回当前会话数量。
	Count() int64
	// Watch 监听首次建立完成的会话；连接迁移成功不会重复通知。
	// 返回的 Signal 在 ctx 取消或 gate 停止后完成。
	Watch(ctx context.Context, handler SessionEstablishedHandler) (async.Signal, error)
}

func newGate(settings ...option.Setting[GateOptions]) IGate {
	return &_Gate{
		options: option.New(With.Default(), settings...),
	}
}

type _Gate struct {
	svcCtx         service.Context
	ctx            context.Context
	terminate      context.CancelCauseFunc
	barrier        generic.Barrier
	options        GateOptions
	tcpListener    net.Listener
	wsListener     *http.Server
	sessions       sync.Map
	sessionCount   atomic.Int64
	sessionWatcher concurrent.Listeners[SessionEstablishedHandler, ISession]
}

// Init 按配置启动 TCP 和 WebSocket 监听器，并异步受理连接以建立会话。
// 两种监听地址均未配置或监听失败时会 panic。
func (g *_Gate) Init(svcCtx service.Context) {
	log.L(svcCtx).Info("initializing add-in", zap.String("name", AddIn.Name))

	g.svcCtx = svcCtx
	g.ctx, g.terminate = context.WithCancelCause(context.Background())

	if g.options.TCPAddress != "" {
		listener, err := newListenConfig(&g.options).Listen(context.Background(), "tcp", g.options.TCPAddress)
		if err != nil {
			log.L(svcCtx).Panic("listen tcp failed", zap.String("address", g.options.TCPAddress), zap.Error(err))
		}

		if g.options.TCPTLSConfig != nil {
			listener = tls.NewListener(listener, g.options.TCPTLSConfig)
		}

		g.tcpListener = listener

		log.L(svcCtx).Info("listener(tcp) started", zap.String("address", listener.Addr().String()))

		go func() {
			for {
				conn, err := g.tcpListener.Accept()
				if err != nil {
					if errors.Is(err, net.ErrClosed) {
						log.L(g.svcCtx).Debug("listener(tcp) closed",
							zap.String("address", g.tcpListener.Addr().String()),
							zap.Error(err))
						return
					}

					log.L(g.svcCtx).Error("listener(tcp) accept a new connection failed",
						zap.String("address", g.tcpListener.Addr().String()),
						zap.Error(err))
					continue
				}

				log.L(g.svcCtx).Debug("listener(tcp) accepted a new connection, establishing session",
					zap.String("address", g.tcpListener.Addr().String()),
					zap.String("remote", conn.RemoteAddr().String()))

				go g.establishSession(conn)
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
				log.L(svcCtx).Panic("listener(ws) need to provide a valid TLS configuration",
					zap.String("address", listener.Addr))
			}
		}

		mux := http.NewServeMux()
		mux.Handle(g.options.WebSocketURL.Path, websocket.Server{
			Handler: func(wsConn *websocket.Conn) {
				wsConn.PayloadType = websocket.BinaryFrame
				conn := &_WebSocketConn{Conn: wsConn, gate: g}

				log.L(g.svcCtx).Debug("listener(ws) accepted a new connection, establishing session",
					zap.String("address", g.wsListener.Addr),
					zap.String("remote", conn.RemoteAddr().String()))

				if session, ok := g.establishSession(conn); ok {
					<-session.Closed().Done()
				}
			},
		})
		listener.Handler = mux

		g.wsListener = listener

		log.L(svcCtx).Info("listener(ws) started", zap.String("address", listener.Addr))

		go func() {
			if err := g.wsListener.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.L(g.svcCtx).Panic("listener(ws) was interrupted", zap.String("address", listener.Addr), zap.Error(err))
			}
		}()
	}

	if g.tcpListener == nil && g.wsListener == nil {
		log.L(svcCtx).Panic("no address need to listen")
	}
}

// Shut 以服务关闭原因为全部会话发起终止，等待受管任务退出后关闭监听器。
func (g *_Gate) Shut(svcCtx service.Context) {
	log.L(svcCtx).Info("shutting down add-in", zap.String("name", AddIn.Name))

	g.terminate(&transport.RstError{
		Code:    gtp.Code_Shutdown,
		Message: "service shutdown",
	})
	g.barrier.Close()
	g.barrier.Wait()

	if g.tcpListener != nil {
		g.tcpListener.Close()
	}
	if g.wsListener != nil {
		g.wsListener.Close()
	}
}

// Get 按会话 ID 查询当前尚未过期的会话。
func (g *_Gate) Get(sessionId uid.Id) (ISession, bool) {
	return g.getSession(sessionId)
}

// Count 返回当前会话数量。
func (g *_Gate) Count() int64 {
	return g.sessionCount.Load()
}

// Watch 监听首次建立完成的会话；连接迁移成功不会重复通知。
// 返回的 Signal 在 ctx 取消或 gate 停止后完成。
func (g *_Gate) Watch(ctx context.Context, handler SessionEstablishedHandler) (async.Signal, error) {
	if handler == nil {
		return async.Signal{}, errors.New("gate: handler is nil")
	}
	stopped, err := g.addSessionWatcher(ctx, handler)
	if err != nil {
		return async.Signal{}, err
	}
	return stopped, nil
}
