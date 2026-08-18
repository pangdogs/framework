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

package gtp

import (
	"errors"
	"io"
)

var (
	// ErrGTP 是 GTP 消息处理错误的根错误。
	ErrGTP = errors.New("gtp")
)

// ReadableMsg 表示可编码到字节流并可复制的 GTP 消息。
type ReadableMsg interface {
	io.Reader
	// Size 返回消息编码后的字节数。
	Size() int
	// MsgID 返回消息类型 ID。
	MsgID() MsgID
	// Clone 返回可独立修改的消息副本。
	Clone() Msg
}

// Msg 表示既可编码也可从字节流解码的 GTP 消息。
type Msg interface {
	ReadableMsg
	io.Writer
}
