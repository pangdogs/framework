package tcp

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

func newTcpGate(options ...GateOption) gate.Gate {
	opts := GateOptions{}
	WithOption{}.Default()(&opts)

	for i := range options {
		options[i](&opts)
	}

	return &_TcpGate{
		options: opts,
	}
}

type _TcpGate struct {
	options    GateOptions
	ctx        service.Context
	listeners  []net.Listener
	sessionMap sync.Map
}

// InitSP 初始化服务插件
func (g *_TcpGate) InitSP(ctx service.Context) {
	logger.Infof(ctx, "init service plugin %q with %q", definePlugin.Name, util.TypeOfAnyFullName(*g))

	g.ctx = ctx

	if len(g.options.Endpoints) <= 0 {
		panic("no endpoints to listen")
	}

	listenConf := g.getListenConfig()

	for _, endpoint := range g.options.Endpoints {
		listener, err := listenConf.Listen(context.Background(), "tcp", endpoint)
		if err != nil {
			logger.Panicf(ctx, "listen %q failed, %s", endpoint, err)
		}

		if g.options.TLSConfig != nil {
			listener = tls.NewListener(listener, g.options.TLSConfig)
		}

		g.listeners = append(g.listeners, listener)
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

				logger.Debugf(ctx, "listener %q accept a new connection, client address %s", listener.Addr(), conn.RemoteAddr())

				go g.newSession(conn)
			}
		}(listener)
	}
}

// ShutSP 关闭服务插件
func (g *_TcpGate) ShutSP(ctx service.Context) {
	logger.Infof(ctx, "shut service plugin %q", definePlugin.Name)

	for _, listener := range g.listeners {
		listener.Close()
	}
}

// Broadcast 广播数据
func (g *_TcpGate) Broadcast(data []byte) error {

}

// Multicast 组播数据
func (g *_TcpGate) Multicast(groupId string, data []byte) error {

}

// Unicast 单播数据
func (g *_TcpGate) Unicast(sessionId string, data []byte) error {

}
