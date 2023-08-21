package gate

import (
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/protocol"
)

type Option struct{}

type (
	StateChangedHandler = func(session Session, old, new SessionState)                     // 会话状态变化的处理器
	RecvDataHandler     = func(session Session, data []byte) error                         // 会话接收的数据的处理器
	RecvEventHandler    = func(session Session, event protocol.Event[transport.Msg]) error // 会话接收的自定义事件的处理器
)

type SessionOptions struct {
	StateChangedHandlers []StateChangedHandler              // 接收会话状态变化的处理器
	RecvDataHandlers     []RecvDataHandler                  // 接收数据的处理器
	RecvEventHandlers    []RecvEventHandler                 // 接收自定义事件的处理器
	SendDataChan         chan []byte                        // 发送数据的channel
	RecvDataChan         chan []byte                        // 接收数据的channel
	SendEventChan        chan protocol.Event[transport.Msg] // 发送自定义事件的channel
	RecvEventChan        chan protocol.Event[transport.Msg] // 接收自定义事件的channel
}

type SessionOption func(options *SessionOptions)

func (Option) Default() SessionOption {
	return func(options *SessionOptions) {
		Option{}.StateChangedHandlers(nil)(options)
		Option{}.RecvDataHandlers(nil)(options)
		Option{}.RecvEventHandlers(nil)(options)
		Option{}.SendDataChanSize(0)(options)
		Option{}.RecvDataChanSize(0)(options)
		Option{}.SendEventChanSize(0)(options)
		Option{}.RecvEventChanSize(0)(options)
	}
}

func (Option) StateChangedHandlers(handlers ...StateChangedHandler) SessionOption {
	return func(options *SessionOptions) {
		options.StateChangedHandlers = handlers
	}
}

func (Option) RecvDataHandlers(handlers ...RecvDataHandler) SessionOption {
	return func(options *SessionOptions) {
		options.RecvDataHandlers = handlers
	}
}

func (Option) RecvEventHandlers(handlers ...RecvEventHandler) SessionOption {
	return func(options *SessionOptions) {
		options.RecvEventHandlers = handlers
	}
}

func (Option) SendDataChanSize(size int) SessionOption {
	return func(options *SessionOptions) {
		if size > 0 {
			options.SendDataChan = make(chan []byte, size)
		} else {
			options.SendDataChan = nil
		}
	}
}

func (Option) RecvDataChanSize(size int) SessionOption {
	return func(options *SessionOptions) {
		if size > 0 {
			options.RecvDataChan = make(chan []byte, size)
		} else {
			options.RecvDataChan = nil
		}
	}
}

func (Option) SendEventChanSize(size int) SessionOption {
	return func(options *SessionOptions) {
		if size > 0 {
			options.SendEventChan = make(chan protocol.Event[transport.Msg], size)
		} else {
			options.SendEventChan = nil
		}
	}
}

func (Option) RecvEventChanSize(size int) SessionOption {
	return func(options *SessionOptions) {
		if size > 0 {
			options.RecvEventChan = make(chan protocol.Event[transport.Msg], size)
		} else {
			options.RecvEventChan = nil
		}
	}
}
