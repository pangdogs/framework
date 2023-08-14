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

// Broadcast 广播数据
func (g *_GtpGate) Broadcast(data []byte) error {
	return nil
}

// Multicast 组播数据
func (g *_GtpGate) Multicast(groupId string, data []byte) error {
	return nil
}

// Unicast 单播数据
func (g *_GtpGate) Unicast(sessionId string, data []byte) error {
	return nil
}
