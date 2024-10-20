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

package gap

import (
	"errors"
	"io"
)

var (
	ErrGAP = errors.New("gap") // 消息协议错误
)

// Msg 消息接口
type Msg interface {
	MsgReader
	MsgWriter
}

// MsgReader 读取消息
type MsgReader interface {
	io.Reader
	// Size 大小
	Size() int
	// MsgId 消息Id
	MsgId() MsgId
}

// MsgWriter 写入消息
type MsgWriter interface {
	io.Writer
	// Size 大小
	Size() int
	// MsgId 消息Id
	MsgId() MsgId
}
