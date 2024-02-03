package gtp_gate

import (
	"context"
	"crypto/tls"
	"errors"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/transport"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/util/concurrent"
	"net"
	"sync"
	"sync/atomic"
)

// IGate 网关
type IGate interface {
	// GetSession 查询会话
	GetSession(sessionId string) (ISession, bool)
	// RangeSessions 遍历所有会话
	RangeSessions(fun func(session ISession) bool)
	// CountSessions 统计所有会话数量
	CountSessions() int
}

func newGate(settings ...option.Setting[GateOptions]) IGate {
	return &_Gate{
		options: option.Make(_GateOption{}.Default(), settings...),
	}
}

type _Gate struct {
	ctx          context.Context
	cancel       context.CancelCauseFunc
	options      GateOptions
	servCtx      service.Context
	wg           sync.WaitGroup
	listeners    []net.Listener
	sessionMap   sync.Map
	sessionCount int64
	futures      concurrent.Futures
}

// InitSP 初始化服务插件
func (g *_Gate) InitSP(ctx service.Context) {
	log.Infof(ctx, "init plugin %q", self.Name)

	g.ctx, g.cancel = context.WithCancelCause(context.Background())
	g.servCtx = ctx

	// 初始化异步模型Future
	g.futures = concurrent.MakeFutures(g.ctx, g.options.FutureTimeout)

	if len(g.options.Endpoints) <= 0 {
		log.Panic(g.servCtx, "no endpoints need to listen")
	}

	listenConf := newListenConfig(&g.options)

	for _, endpoint := range g.options.Endpoints {
		listener, err := listenConf.Listen(context.Background(), "tcp", endpoint)
		if err != nil {
			log.Panicf(g.servCtx, "listen %q failed, %s", endpoint, err)
		}

		if g.options.TLSConfig != nil {
			listener = tls.NewListener(listener, g.options.TLSConfig)
		}

		g.listeners = append(g.listeners, listener)

		log.Infof(g.servCtx, "listener %q started", listener.Addr())
	}

	for _, listener := range g.listeners {
		go func(listener net.Listener) {
			for {
				conn, err := listener.Accept()
				if err != nil {
					if errors.Is(err, net.ErrClosed) {
						log.Debugf(g.servCtx, "listener %q closed", listener.Addr())
						return
					}
					log.Errorf(g.servCtx, "listener %q accept a new connection failed, %s", listener.Addr(), err)
					continue
				}

				log.Debugf(g.servCtx, "listener %q accept a new connection, client %q", listener.Addr(), conn.RemoteAddr())

				go g.handleSession(conn)
			}
		}(listener)
	}
}

// ShutSP 关闭服务插件
func (g *_Gate) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut plugin %q", self.Name)

	g.cancel(&transport.RstError{
		Code:    gtp.Code_Shutdown,
		Message: "service shutdown",
	})
	g.wg.Wait()

	for _, listener := range g.listeners {
		listener.Close()
	}
}

// GetSession 查询会话
func (g *_Gate) GetSession(sessionId string) (ISession, bool) {
	return g.loadSession(sessionId)
}

// RangeSessions 遍历所有会话
func (g *_Gate) RangeSessions(fun func(session ISession) bool) {
	if fun == nil {
		return
	}
	g.sessionMap.Range(func(k, v any) bool {
		return fun(v.(ISession))
	})
}

// CountSessions 统计所有会话数量
func (g *_Gate) CountSessions() int {
	return int(atomic.LoadInt64(&g.sessionCount))
}
