package transport

import (
	"git.golaxy.org/plugins/gtp"
)

// Event 消息事件
type Event[T gtp.MsgReader] struct {
	Flags gtp.Flags // 标志位
	Seq   uint32    // 消息序号
	Ack   uint32    // 应答序号
	Msg   T         // 消息
}

// Pack 打包事件，用于发送消息
func (e Event[T]) Pack() Event[gtp.MsgReader] {
	return Event[gtp.MsgReader]{
		Flags: e.Flags,
		Seq:   e.Seq,
		Ack:   e.Ack,
		Msg:   e.Msg,
	}
}

// UnpackEvent 解包事件，用于解析消息
func UnpackEvent[T gtp.MsgReader](me Event[gtp.Msg]) Event[T] {
	e := Event[T]{
		Flags: me.Flags,
		Seq:   me.Seq,
		Ack:   me.Ack,
	}

	if me.Msg == nil {
		return e
	}

	msg, ok := any(me.Msg).(*T)
	if !ok {
		panic("gtp: incorrect msg type")
	}
	e.Msg = *msg

	return e
}
