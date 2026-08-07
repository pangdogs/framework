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
	"io"

	"git.golaxy.org/framework/utils/binaryutil"
)

// MsgSigned 包装原始消息体及其消息认证码。
type MsgSigned struct {
	Data []byte // 被认证数据；解码时引用输入缓冲区。
	MAC  []byte // 消息认证码；解码时引用输入缓冲区。
}

// Read 将认证包装编码到 p。
func (m MsgSigned) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteBytes(m.Data); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteBytes(m.MAC); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write 从 p 解码认证包装。
func (m *MsgSigned) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	var err error

	m.Data, err = bs.ReadBytesRef()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.MAC, err = bs.ReadBytesRef()
	if err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 返回认证包装编码后的字节数。
func (m MsgSigned) Size() int {
	return binaryutil.SizeofBytes(m.Data) + binaryutil.SizeofBytes(m.MAC)
}
