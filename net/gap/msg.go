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
	// ErrGAP 是 GAP 消息处理错误的根错误。
	ErrGAP = errors.New("gap")
)

// ReadableMsg 表示可编码到字节流的 GAP 消息。
type ReadableMsg interface {
	io.Reader
	// Size 返回消息编码后的字节数。
	Size() int
	// MsgId 返回消息类型 ID。
	MsgId() MsgId
}

// Msg 表示既可编码也可从字节流解码的 GAP 消息。
type Msg interface {
	ReadableMsg
	io.Writer
}
