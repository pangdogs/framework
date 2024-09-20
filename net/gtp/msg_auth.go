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
	"bytes"
	"git.golaxy.org/framework/utils/binaryutil"
	"io"
	"strings"
)

// MsgAuth 鉴权（注意：为了提高解码性能，减少内存碎片，解码string与bytes字段时均使用引用类型，引用字节池中的bytes，GC时会被归还字节池，不要直接持有此类型字段）
type MsgAuth struct {
	UserId     string // 用户Id
	Token      string // 令牌
	Extensions []byte // 扩展内容
}

// Read implements io.Reader
func (m MsgAuth) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteString(m.UserId); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteString(m.Token); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteBytes(m.Extensions); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write implements io.Writer
func (m *MsgAuth) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	var err error

	m.UserId, err = bs.ReadStringRef()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.Token, err = bs.ReadStringRef()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.Extensions, err = bs.ReadBytesRef()
	if err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 大小
func (m MsgAuth) Size() int {
	return binaryutil.SizeofString(m.UserId) + binaryutil.SizeofString(m.Token) + binaryutil.SizeofBytes(m.Extensions)
}

// MsgId 消息Id
func (MsgAuth) MsgId() MsgId {
	return MsgId_Auth
}

// Clone 克隆消息对象
func (m MsgAuth) Clone() MsgReader {
	return MsgAuth{
		UserId:     strings.Clone(m.UserId),
		Token:      strings.Clone(m.Token),
		Extensions: bytes.Clone(m.Extensions),
	}
}
