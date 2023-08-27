package gtp_gate

import (
	"kit.golaxy.org/plugins/gtp"
	"kit.golaxy.org/plugins/gtp/transport"
)

type _SessionOption struct{}

type (
	StateChangedHandler = func(old, new SessionState)                // 会话状态变化的处理器
	RecvDataHandler     = func(data []byte) error                    // 会话接收的数据的处理器
	RecvEventHandler    = func(event transport.Event[gtp.Msg]) error // 会话接收的自定义事件的处理器
)

type SessionOptions struct {
	StateChangedHandlers []StateChangedHandler         // 接收会话状态变化的处理器
	RecvDataHandlers     []RecvDataHandler             // 接收数据的处理器
	RecvEventHandlers    []RecvEventHandler            // 接收自定义事件的处理器
	SendDataChan         chan []byte                   // 发送数据的channel
	RecvDataChan         chan []byte                   // 接收数据的channel
	SendEventChan        chan transport.Event[gtp.Msg] // 发送自定义事件的channel
	RecvEventChan        chan transport.Event[gtp.Msg] // 接收自定义事件的channel
}

type SessionOption func(options *SessionOptions)

func (_SessionOption) Default() SessionOption {
	return func(options *SessionOptions) {
		_SessionOption{}.StateChangedHandlers(nil)(options)
		_SessionOption{}.RecvDataHandlers(nil)(options)
		_SessionOption{}.RecvEventHandlers(nil)(options)
		_SessionOption{}.SendDataChanSize(0)(options)
		_SessionOption{}.RecvDataChanSize(0)(options)
		_SessionOption{}.SendEventChanSize(0)(options)
		_SessionOption{}.RecvEventChanSize(0)(options)
	}
}

func (_SessionOption) StateChangedHandlers(handlers ...StateChangedHandler) SessionOption {
	return func(options *SessionOptions) {
		options.StateChangedHandlers = handlers
	}
}

func (_SessionOption) RecvDataHandlers(handlers ...RecvDataHandler) SessionOption {
	return func(options *SessionOptions) {
		options.RecvDataHandlers = handlers
	}
}

func (_SessionOption) RecvEventHandlers(handlers ...RecvEventHandler) SessionOption {
	return func(options *SessionOptions) {
		options.RecvEventHandlers = handlers
	}
}

func (_SessionOption) SendDataChanSize(size int) SessionOption {
	return func(options *SessionOptions) {
		if size > 0 {
			options.SendDataChan = make(chan []byte, size)
		} else {
			options.SendDataChan = nil
		}
	}
}

func (_SessionOption) RecvDataChanSize(size int) SessionOption {
	return func(options *SessionOptions) {
		if size > 0 {
			options.RecvDataChan = make(chan []byte, size)
		} else {
			options.RecvDataChan = nil
		}
	}
}

func (_SessionOption) SendEventChanSize(size int) SessionOption {
	return func(options *SessionOptions) {
		if size > 0 {
			options.SendEventChan = make(chan transport.Event[gtp.Msg], size)
		} else {
			options.SendEventChan = nil
		}
	}
}

func (_SessionOption) RecvEventChanSize(size int) SessionOption {
	return func(options *SessionOptions) {
		if size > 0 {
			options.RecvEventChan = make(chan transport.Event[gtp.Msg], size)
		} else {
			options.RecvEventChan = nil
		}
	}
}
