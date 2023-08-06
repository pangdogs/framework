package gate

import (
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/protocol"
)

type (
	StateChangedHandler = func(session Session, old, new SessionState)                     // 会话状态变化的处理器
	RecvDataHandler     = func(session Session, data []byte, sequenced bool) error         // 会话接收的数据的处理器
	RecvEventHandler    = func(session Session, event protocol.Event[transport.Msg]) error // 会话接收的自定义事件的处理器
)

// SessionSetting 会话设置接口（在会话状态为握手中时可用）
type SessionSetting interface {
	// InitStateChangedHandlers 设置接收会话状态变化的处理器
	InitStateChangedHandlers(handlers []StateChangedHandler) error
	// InitRecvDataHandlers 设置接收数据的处理器
	InitRecvDataHandlers(handlers []RecvDataHandler) error
	// InitRecvEventHandlers 设置接收自定义事件的处理器
	InitRecvEventHandlers(handlers []RecvEventHandler) error
	// InitRecvDataChanSize 设置接收数据的chan的大小，<=0表示不使用chan
	InitRecvDataChanSize(size int) error
	// InitRecvEventSize 设置自定义事件的chan的大小，<=0表示不使用chan
	InitRecvEventSize(size int) error
}
