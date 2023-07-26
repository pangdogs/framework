package gate

import (
	"kit.golaxy.org/plugins/transport/protocol"
)

type (
	StateChangedHandler = func(old, new SessionState)             // 会话状态变化的处理器
	RecvHandler         = func(data []byte, sequenced bool) error // 接收的数据的处理器
	RecvEventHandler    = protocol.EventHandler                   // 接收的自定义事件的处理器
)

// SessionHandshake 会话握手状态接口
type SessionHandshake interface {
	// StateChangedHandler 设置接收会话状态变化的处理器
	StateChangedHandler(handler StateChangedHandler) error
	// RecvHandler 设置接收数据的处理器
	RecvHandler(handler RecvHandler) error
	// RecvChanSize 设置接收数据的chan的大小，<=0表示不使用chan
	RecvChanSize(size int) error
	// RecvEventHandlers 设置接收自定义事件的处理器
	RecvEventHandlers(handlers []RecvEventHandler) error
	// RecvEventSize 设置自定义事件的chan的大小，<=0表示不使用chan
	RecvEventSize(size int) error
}
