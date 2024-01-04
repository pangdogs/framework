package gtp_gate

import (
	"context"
	"crypto/tls"
	"errors"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util/option"
	"kit.golaxy.org/golaxy/util/types"
	"kit.golaxy.org/plugins/log"
	"kit.golaxy.org/plugins/util/concurrent"
	"net"
	"sync"
	"sync/atomic"
)

// Gate 网关
type Gate interface {
	// GetSession 查询会话
	GetSession(sessionId string) (Session, bool)
	// RangeSessions 遍历所有会话
	RangeSessions(fun func(session Session) bool)
	// CountSessions 统计所有会话数量
	CountSessions() int
}

func newGate(settings ...option.Setting[GateOptions]) Gate {
	return &_Gate{
		options: option.Make(_GateOption{}.Default(), settings...),
	}
}

type _Gate struct {
	options      GateOptions
	ctx          service.Context
	wg           sync.WaitGroup
	listeners    []net.Listener
	sessionMap   sync.Map
	sessionCount int64
	futures      concurrent.Futures
}

// InitSP 初始化服务插件
func (g *_Gate) InitSP(ctx service.Context) {
	log.Infof(ctx, "init service plugin <%s>:%s", plugin.Name, types.AnyFullName(*g))

	g.ctx = ctx
	g.futures = concurrent.MakeFutures(ctx, g.options.FutureTimeout)

	if len(g.options.Endpoints) <= 0 {
		log.Panic(ctx, "no endpoints need to listen")
	}

	listenConf := newListenConfig(&g.options)

	for _, endpoint := range g.options.Endpoints {
		listener, err := listenConf.Listen(context.Background(), "tcp", endpoint)
		if err != nil {
			log.Panicf(ctx, "listen %q failed, %s", endpoint, err)
		}

		if g.options.TLSConfig != nil {
			listener = tls.NewListener(listener, g.options.TLSConfig)
		}

		g.listeners = append(g.listeners, listener)

		log.Infof(g.ctx, "listener %q started", listener.Addr())
	}

	for _, listener := range g.listeners {
		g.wg.Add(1)
		go func(listener net.Listener) {
			defer g.wg.Done()
			for {
				conn, err := listener.Accept()
				if err != nil {
					if errors.Is(err, net.ErrClosed) {
						log.Debugf(ctx, "listener %q closed", listener.Addr())
						return
					}
					log.Errorf(ctx, "listener %q accept a new connection failed, %s", listener.Addr(), err)
					continue
				}

				log.Debugf(ctx, "listener %q accept a new connection, client %q", listener.Addr(), conn.RemoteAddr())

				go g.handleSession(conn)
			}
		}(listener)
	}
}

// ShutSP 关闭服务插件
func (g *_Gate) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut service plugin <%s>:%s", plugin.Name, types.AnyFullName(*g))

	g.wg.Wait()

	for _, listener := range g.listeners {
		listener.Close()
	}
}

// GetSession 查询会话
func (g *_Gate) GetSession(sessionId string) (Session, bool) {
	return g.loadSession(sessionId)
}

// RangeSessions 遍历所有会话
func (g *_Gate) RangeSessions(fun func(session Session) bool) {
	if fun == nil {
		return
	}
	g.sessionMap.Range(func(k, v any) bool {
		return fun(v.(Session))
	})
}

// CountSessions 统计所有会话数量
func (g *_Gate) CountSessions() int {
	return int(atomic.LoadInt64(&g.sessionCount))
}
