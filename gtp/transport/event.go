package transport

import (
	"fmt"
	"kit.golaxy.org/plugins/gtp"
)

// Event 消息事件
type Event[T gtp.Msg] struct {
	Flags gtp.Flags // 标志位
	Seq   uint32    // 消息序号
	Ack   uint32    // 应答序号
	Msg   T         // 消息
}

// Clone 克隆消息事件
func (e Event[T]) Clone() Event[T] {
	e.Msg = e.Msg.Clone().(T)
	return e
}

// UnpackEvent 解包消息事件
func UnpackEvent[T gtp.Msg](pe Event[gtp.Msg]) Event[T] {
	return Event[T]{
		Flags: pe.Flags,
		Seq:   pe.Seq,
		Ack:   pe.Ack,
		Msg:   pe.Msg.(T),
	}
}

// PackEvent 打包消息事件
func PackEvent[T gtp.Msg](e Event[T]) Event[gtp.Msg] {
	return Event[gtp.Msg]{
		Flags: e.Flags,
		Seq:   e.Seq,
		Ack:   e.Ack,
		Msg:   e.Msg,
	}
}

// RstError Rst错误提示
type RstError struct {
	Code    gtp.Code // 错误码
	Message string   // 错误信息
}

// Error 错误信息
func (e *RstError) Error() string {
	return fmt.Sprintf("(%d) %s", e.Code, e.Message)
}

// EventToRstErr Rst错误消息事件转换为错误提示
func EventToRstErr(e Event[*gtp.MsgRst]) *RstError {
	return &RstError{
		Code:    e.Msg.Code,
		Message: e.Msg.Message,
	}
}

// RstErrToEvent Rst错误提示转换为消息事件
func RstErrToEvent(err *RstError) Event[*gtp.MsgRst] {
	return Event[*gtp.MsgRst]{
		Msg: &gtp.MsgRst{
			Code:    err.Code,
			Message: err.Message,
		},
	}
}
