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

	"git.golaxy.org/framework/utils/binaryutil"
)

// MsgPayload 携带上层业务数据。直接通过 Write 或 Unmarshal 解码时，Data 会引用输入切片；
// 输入将被复用或修改时应先 Clone。Decoder.Decode 返回的消息不引用调用方输入。
type MsgPayload struct {
	Data []byte // 业务负载。
}

// Read 将业务负载编码到 p。
func (m MsgPayload) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteBytes(m.Data); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write 从 p 解码业务负载，Data 会引用 p。
func (m *MsgPayload) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	var err error

	m.Data, err = bs.ReadBytesRef()
	if err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 返回业务负载编码后的字节数。
func (m MsgPayload) Size() int {
	return binaryutil.SizeofBytes(m.Data)
}

// MsgID 返回业务负载消息的内置类型 ID。
func (MsgPayload) MsgID() MsgID {
	return MsgID_Payload
}

// Clone 深复制业务负载。
func (m MsgPayload) Clone() Msg {
	return &MsgPayload{
		Data: bytes.Clone(m.Data),
	}
}
