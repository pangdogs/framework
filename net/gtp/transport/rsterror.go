/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

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

// ToEvent 转换为消息事件
func (err RstError) ToEvent() Event[*gtp.MsgRst] {
	return Event[*gtp.MsgRst]{
		Msg: &gtp.MsgRst{
			Code:    err.Code,
			Message: err.Message,
		},
	}
}

// CastRstErr Rst错误事件转换为错误提示
func CastRstErr(e Event[*gtp.MsgRst]) *RstError {
	return &RstError{
		Code:    e.Msg.Code,
		Message: e.Msg.Message,
	}
}
