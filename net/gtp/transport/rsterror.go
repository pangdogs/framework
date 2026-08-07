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
	"strings"

	"git.golaxy.org/framework/net/gtp"
)

// RstError 将 GTP 链路重置码和消息表示为 Go error。
type RstError struct {
	Code    gtp.Code // 链路重置原因码。
	Message string   // 可读错误消息。
}

// Error 返回包含重置码和消息的文本。
func (err RstError) Error() string {
	return fmt.Sprintf("(%d) %s", err.Code, err.Message)
}

// ToEvent 将错误转换为链路重置事件。
func (err RstError) ToEvent() Event[*gtp.MsgRst] {
	return Event[*gtp.MsgRst]{
		Msg: &gtp.MsgRst{
			Code:    err.Code,
			Message: err.Message,
		},
	}
}

// ToRstError 将链路重置事件转换为独立持有消息文本的错误。
func ToRstError(e Event[*gtp.MsgRst]) *RstError {
	return &RstError{
		Code:    e.Msg.Code,
		Message: strings.Clone(e.Msg.Message),
	}
}
