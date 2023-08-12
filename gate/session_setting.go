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

// SessionSetting 会话设置接口（在会话状态Handshake与Confirmed时可用）
type SessionSetting interface {
	// StateChangedHandlers 设置接收会话状态变化的处理器
	StateChangedHandlers(handlers []StateChangedHandler) error
	// RecvDataHandlers 设置接收数据的处理器
	RecvDataHandlers(handlers []RecvDataHandler) error
	// RecvEventHandlers 设置接收自定义事件的处理器
	RecvEventHandlers(handlers []RecvEventHandler) error
	// SendDataChanSize 设置发送数据的chan的大小，<=0表示不使用channel
	SendDataChanSize(size int) error
	// RecvDataChanSize 设置接收数据的chan的大小，<=0表示不使用channel
	RecvDataChanSize(size int) error
	// SendEventSize 设置发送自定义事件的chan的大小，<=0表示不使用channel
	SendEventSize(size int) error
	// RecvEventSize 设置接收自定义事件的chan的大小，<=0表示不使用channel
	RecvEventSize(size int) error
}
