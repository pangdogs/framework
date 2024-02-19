package gate

import (
	"git.golaxy.org/core/util/generic"
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/transport"
)

type (
	StateChangedHandler = generic.DelegateAction2[SessionState, SessionState]    // 会话状态变化的处理器
	RecvDataHandler     = generic.DelegateFunc1[[]byte, error]                   // 会话接收的数据的处理器
	RecvEventHandler    = generic.DelegateFunc1[transport.Event[gtp.Msg], error] // 会话接收的自定义事件的处理器
)

type SessionOptions struct {
	StateChangedHandler StateChangedHandler                 // 接收会话状态变化的处理器
	RecvDataHandler     RecvDataHandler                     // 接收数据的处理器（优先级低于监控器）
	RecvEventHandler    RecvEventHandler                    // 接收自定义事件的处理器（优先级低于监控器）
	SendDataChan        chan []byte                         // 发送数据的channel
	RecvDataChan        chan []byte                         // 接收数据的channel
	SendEventChan       chan transport.Event[gtp.MsgReader] // 发送自定义事件的channel
	RecvEventChan       chan transport.Event[gtp.Msg]       // 接收自定义事件的channel
}

type _SessionOption struct{}

func (_SessionOption) Default() option.Setting[SessionOptions] {
	return func(options *SessionOptions) {
		With.Session.StateChangedHandler(nil)(options)
		With.Session.RecvDataHandler(nil)(options)
		With.Session.RecvEventHandler(nil)(options)
		With.Session.SendDataChanSize(0)(options)
		With.Session.RecvDataChanSize(0)(options)
		With.Session.SendEventChanSize(0)(options)
		With.Session.RecvEventChanSize(0)(options)
	}
}

func (_SessionOption) StateChangedHandler(handler StateChangedHandler) option.Setting[SessionOptions] {
	return func(options *SessionOptions) {
		options.StateChangedHandler = handler
	}
}

func (_SessionOption) RecvDataHandler(handler RecvDataHandler) option.Setting[SessionOptions] {
	return func(options *SessionOptions) {
		options.RecvDataHandler = handler
	}
}

func (_SessionOption) RecvEventHandler(handler RecvEventHandler) option.Setting[SessionOptions] {
	return func(options *SessionOptions) {
		options.RecvEventHandler = handler
	}
}

func (_SessionOption) SendDataChanSize(size int) option.Setting[SessionOptions] {
	return func(options *SessionOptions) {
		if size > 0 {
			options.SendDataChan = make(chan []byte, size)
		} else {
			options.SendDataChan = nil
		}
	}
}

func (_SessionOption) RecvDataChanSize(size int) option.Setting[SessionOptions] {
	return func(options *SessionOptions) {
		if size > 0 {
			options.RecvDataChan = make(chan []byte, size)
		} else {
			options.RecvDataChan = nil
		}
	}
}

func (_SessionOption) SendEventChanSize(size int) option.Setting[SessionOptions] {
	return func(options *SessionOptions) {
		if size > 0 {
			options.SendEventChan = make(chan transport.Event[gtp.MsgReader], size)
		} else {
			options.SendEventChan = nil
		}
	}
}

func (_SessionOption) RecvEventChanSize(size int) option.Setting[SessionOptions] {
	return func(options *SessionOptions) {
		if size > 0 {
			options.RecvEventChan = make(chan transport.Event[gtp.Msg], size)
		} else {
			options.RecvEventChan = nil
		}
	}
}
