package gate

import (
	"context"
	"crypto/tls"
	"errors"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/transport"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/util/concurrent"
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
	ctx             context.Context
	terminate       context.CancelCauseFunc
	options         GateOptions
	servCtx         service.Context
	wg              sync.WaitGroup
	tcpListener     net.Listener
	wsListener      *http.Server
	sessionMap      sync.Map
	sessionCount    int64
	sessionWatchers concurrent.LockedSlice[*_SessionWatcher]
}

// InitSP 初始化服务插件
func (g *_Gate) InitSP(ctx service.Context) {
	log.Infof(ctx, "init plugin %q", self.Name)

	g.ctx, g.terminate = context.WithCancelCause(context.Background())
	g.servCtx = ctx

	if g.options.TCPAddress != "" {
		listener, err := newListenConfig(&g.options).Listen(context.Background(), "tcp", g.options.TCPAddress)
		if err != nil {
			log.Panicf(g.servCtx, "listen tcp %q failed, %s", g.options.TCPAddress, err)
		}

		if g.options.TCPTLSConfig != nil {
			listener = tls.NewListener(listener, g.options.TCPTLSConfig)
		}

		g.tcpListener = listener

		log.Infof(g.servCtx, "listener %q started", g.tcpListener.Addr())

		go func() {
			for {
				conn, err := g.tcpListener.Accept()
				if err != nil {
					if errors.Is(err, net.ErrClosed) {
						log.Debugf(g.servCtx, "listener %q closed", g.tcpListener.Addr())
						return
					}
					log.Errorf(g.servCtx, "listener %q accept a new connection failed, %s", g.tcpListener.Addr(), err)
					continue
				}

				log.Debugf(g.servCtx, "listener %q accept a new connection, remote %q", g.tcpListener.Addr(), conn.RemoteAddr())
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
				log.Panicf(g.servCtx, "use HTTPS to listen, need to provide a valid TLS configuration")
			}
		}

		mux := http.NewServeMux()
		mux.Handle(g.options.WebSocketURL.Path, websocket.Handler(func(conn *websocket.Conn) {
			log.Debugf(g.servCtx, "listener %q accept a new connection, remote %q", conn.LocalAddr(), conn.RemoteAddr())
			if session, ok := g.handleSession(conn); ok {
				<-session.Done()
				<-session.Close(nil)
			}
		}))
		listener.Handler = mux

		g.wsListener = listener

		log.Infof(g.servCtx, "listener %q started", g.options.WebSocketURL)

		go func() {
			if err := g.wsListener.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Panicf(g.servCtx, "listener %q was interrupted, %s", g.options.WebSocketURL, err)
			}
		}()
	}

	if g.tcpListener == nil && g.wsListener == nil {
		log.Panic(g.servCtx, "no address need to listen")
	}
}

// ShutSP 关闭服务插件
func (g *_Gate) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut plugin %q", self.Name)

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
