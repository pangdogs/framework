package transport

import (
	"fmt"
	"kit.golaxy.org/plugins/gtp"
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

// Event 转换为消息事件
func (err RstError) Event() Event[gtp.MsgRst] {
	return Event[gtp.MsgRst]{
		Msg: gtp.MsgRst{
			Code:    err.Code,
			Message: err.Message,
		},
	}
}

// EventRstToRstErr Rst错误事件转换为错误提示
func EventRstToRstErr(e Event[gtp.MsgRst]) RstError {
	return RstError{
		Code:    e.Msg.Code,
		Message: e.Msg.Message,
	}
}
