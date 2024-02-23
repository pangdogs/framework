package gate

import (
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/transport"
)

type _SessionOptions struct {
	StateChangedHandler SessionStateChangedHandler          // 会话状态变化的处理器
	SendDataChan        chan []byte                         // 发送数据的channel
	RecvDataChan        chan []byte                         // 接收数据的channel
	SendEventChan       chan transport.Event[gtp.MsgReader] // 发送自定义事件的channel
	RecvEventChan       chan transport.Event[gtp.Msg]       // 接收自定义事件的channel
	RecvDataHandler     SessionRecvDataHandler              // 接收数据的处理器（优先级低于监控器）
	RecvEventHandler    SessionRecvEventHandler             // 接收自定义事件的处理器（优先级低于监控器）
}

var sessionWith _SessionOption

type _SessionOption struct{}

func (_SessionOption) Default() option.Setting[_SessionOptions] {
	return func(options *_SessionOptions) {
		sessionWith.StateChangedHandler(nil)(options)
		sessionWith.SendDataChanSize(0)(options)
		sessionWith.RecvDataChanSize(0)(options)
		sessionWith.SendEventChanSize(0)(options)
		sessionWith.RecvEventChanSize(0)(options)
		sessionWith.RecvDataHandler(nil)(options)
		sessionWith.RecvEventHandler(nil)(options)
	}
}

func (_SessionOption) StateChangedHandler(handler SessionStateChangedHandler) option.Setting[_SessionOptions] {
	return func(options *_SessionOptions) {
		options.StateChangedHandler = handler
	}
}

func (_SessionOption) SendDataChanSize(size int) option.Setting[_SessionOptions] {
	return func(options *_SessionOptions) {
		if size > 0 {
			options.SendDataChan = make(chan []byte, size)
		} else {
			options.SendDataChan = nil
		}
	}
}

func (_SessionOption) RecvDataChanSize(size int) option.Setting[_SessionOptions] {
	return func(options *_SessionOptions) {
		if size > 0 {
			options.RecvDataChan = make(chan []byte, size)
		} else {
			options.RecvDataChan = nil
		}
	}
}

func (_SessionOption) SendEventChanSize(size int) option.Setting[_SessionOptions] {
	return func(options *_SessionOptions) {
		if size > 0 {
			options.SendEventChan = make(chan transport.Event[gtp.MsgReader], size)
		} else {
			options.SendEventChan = nil
		}
	}
}

func (_SessionOption) RecvEventChanSize(size int) option.Setting[_SessionOptions] {
	return func(options *_SessionOptions) {
		if size > 0 {
			options.RecvEventChan = make(chan transport.Event[gtp.Msg], size)
		} else {
			options.RecvEventChan = nil
		}
	}
}

func (_SessionOption) RecvDataHandler(handler SessionRecvDataHandler) option.Setting[_SessionOptions] {
	return func(options *_SessionOptions) {
		options.RecvDataHandler = handler
	}
}

func (_SessionOption) RecvEventHandler(handler SessionRecvEventHandler) option.Setting[_SessionOptions] {
	return func(options *_SessionOptions) {
		options.RecvEventHandler = handler
	}
}
