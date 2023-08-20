package gtp

import (
	"context"
	"crypto/tls"
	"errors"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util"
	"kit.golaxy.org/plugins/gate"
	"kit.golaxy.org/plugins/logger"
	"net"
	"sync"
	"sync/atomic"
)

func newGtpGate(options ...GateOption) gate.Gate {
	opts := GateOptions{}
	Option{}.Default()(&opts)

	for i := range options {
		options[i](&opts)
	}

	return &_GtpGate{
		options: opts,
	}
}

type _GtpGate struct {
	options      GateOptions
	ctx          service.Context
	listeners    []net.Listener
	sessionMap   sync.Map
	sessionCount int64
}

// InitSP 初始化服务插件
func (g *_GtpGate) InitSP(ctx service.Context) {
	logger.Infof(ctx, "init service plugin %q with %q", definePlugin.Name, util.TypeOfAnyFullName(*g))

	g.ctx = ctx

	if len(g.options.Endpoints) <= 0 {
		logger.Panic(ctx, "no endpoints need to listen")
	}

	listenConf := newListenConfig(&g.options)

	for _, endpoint := range g.options.Endpoints {
		listener, err := listenConf.Listen(context.Background(), "tcp", endpoint)
		if err != nil {
			logger.Panicf(ctx, "listen %q failed, %s", endpoint, err)
		}

		if g.options.TLSConfig != nil {
			listener = tls.NewListener(listener, g.options.TLSConfig)
		}

		g.listeners = append(g.listeners, listener)

		logger.Infof(g.ctx, "listener %q started", listener.Addr())
	}

	for _, listener := range g.listeners {
		go func(listener net.Listener) {
			for {
				conn, err := listener.Accept()
				if err != nil {
					if errors.Is(err, net.ErrClosed) {
						logger.Debugf(ctx, "listener %q closed", listener.Addr())
						return
					}
					logger.Errorf(ctx, "listener %q accept a new connection failed, %s", listener.Addr(), err)
					continue
				}

				logger.Debugf(ctx, "listener %q accept a new connection, client %q", listener.Addr(), conn.RemoteAddr())

				go g.HandleSession(conn)
			}
		}(listener)
	}
}

// ShutSP 关闭服务插件
func (g *_GtpGate) ShutSP(ctx service.Context) {
	logger.Infof(ctx, "shut service plugin %q", definePlugin.Name)

	for _, listener := range g.listeners {
		listener.Close()
	}
}

// GetSession 查询会话
func (g *_GtpGate) GetSession(sessionId string) (gate.Session, bool) {
	return g.LoadSession(sessionId)
}

// RangeSessions 遍历所有会话
func (g *_GtpGate) RangeSessions(fun func(session gate.Session) bool) {
	if fun == nil {
		return
	}
	g.sessionMap.Range(func(k, v any) bool {
		return fun(v.(gate.Session))
	})
}

// CountSessions 统计所有会话数量
func (g *_GtpGate) CountSessions() int {
	return int(atomic.LoadInt64(&g.sessionCount))
}
