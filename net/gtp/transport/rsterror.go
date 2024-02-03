package transport

import (
	"fmt"
	"git.golaxy.org/framework/net/gtp"
)

// RstError Rst错误提示
type RstError struct {
	Code    gtp.Code // 错误码
	Message string   // 错误信息
}

// Error 错误信息
func (err RstError) Error() string {
	return fmt.Sprintf("(%d) %s", err.Code, err.Message)
}

// CastEvent 转换为消息事件
func (err RstError) CastEvent() Event[gtp.MsgRst] {
	return Event[gtp.MsgRst]{
		Msg: gtp.MsgRst{
			Code:    err.Code,
			Message: err.Message,
		},
	}
}

// CastRstErr Rst错误事件转换为错误提示
func CastRstErr(e Event[gtp.MsgRst]) RstError {
	return RstError{
		Code:    e.Msg.Code,
		Message: e.Msg.Message,
	}
}
