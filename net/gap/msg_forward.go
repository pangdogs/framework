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
	"io"

	"git.golaxy.org/framework/utils/binaryutil"
)

// MsgForward 将一个已编码 GAP 消息路由到目标地址。
type MsgForward struct {
	Src       Origin // 原始消息来源。
	Dst       string // 目标通信地址。
	CorrId    int64  // 请求关联 ID；无需关联时为零。
	TransId   MsgId  // 被转发消息的类型 ID。
	TransData []byte // 被转发消息的已编码内容；解码时引用输入缓冲区。
}

// Read 将转发消息编码到 p。
func (m MsgForward) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if _, err := binaryutil.CopyToByteStream(&bs, m.Src); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteString(m.Dst); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteVarint(m.CorrId); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteUint32(m.TransId); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteBytes(m.TransData); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write 从 p 解码转发消息。
func (m *MsgForward) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	var err error

	_, err = bs.WriteTo(&m.Src)
	if err != nil {
		return bs.BytesRead(), err
	}

	m.Dst, err = bs.ReadString()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.CorrId, err = bs.ReadVarint()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.TransId, err = bs.ReadUint32()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.TransData, err = bs.ReadBytesRef()
	if err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 返回转发消息编码后的字节数。
func (m MsgForward) Size() int {
	return m.Src.Size() + binaryutil.SizeofString(m.Dst) + binaryutil.SizeofVarint(m.CorrId) + binaryutil.SizeofUint32 + binaryutil.SizeofBytes(m.TransData)
}

// MsgId 返回转发消息的内置类型 ID。
func (MsgForward) MsgId() MsgId {
	return MsgId_Forward
}
