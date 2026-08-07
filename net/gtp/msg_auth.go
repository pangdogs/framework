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
	"io"
	"strings"

	"git.golaxy.org/framework/utils/binaryutil"
)

// MsgAuth 携带客户端鉴权信息。直接通过 Write 或 Unmarshal 解码时，字段会引用输入切片；
// 输入将被复用或修改时应先 Clone。Decoder.Decode 返回的消息不引用调用方输入。
type MsgAuth struct {
	UserId     string // 用户 ID。
	Token      string // 鉴权令牌。
	Extensions []byte // 业务扩展数据。
}

// Read 将鉴权消息编码到 p。
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

// Write 从 p 解码鉴权消息，字段会引用 p。
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

// Size 返回鉴权消息编码后的字节数。
func (m MsgAuth) Size() int {
	return binaryutil.SizeofString(m.UserId) + binaryutil.SizeofString(m.Token) + binaryutil.SizeofBytes(m.Extensions)
}

// MsgId 返回鉴权消息的内置类型 ID。
func (MsgAuth) MsgId() MsgId {
	return MsgId_Auth
}

// Clone 深复制所有引用型字段。
func (m MsgAuth) Clone() Msg {
	return &MsgAuth{
		UserId:     strings.Clone(m.UserId),
		Token:      strings.Clone(m.Token),
		Extensions: bytes.Clone(m.Extensions),
	}
}
