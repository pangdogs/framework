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

const (
	// Flag_VerifyEncryption 表示密码规范切换消息携带用于验证加密结果的数据。
	Flag_VerifyEncryption Flag = 1 << (iota + Flag_Customize)
)

// MsgChangeCipherSpec 通知对端启用协商后的密码规范。直接通过 Write 或 Unmarshal 解码时，
// EncryptedHello 会引用输入切片；输入将被复用或修改时应先 Clone。Decoder.Decode 返回的消息不引用调用方输入。
type MsgChangeCipherSpec struct {
	EncryptedHello []byte // 可选的加密 Hello，用于双方验证加密是否成功。
}

// Read 将密码规范切换消息编码到 p。
func (m MsgChangeCipherSpec) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteBytes(m.EncryptedHello); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write 从 p 解码密码规范切换消息，字段会引用 p。
func (m *MsgChangeCipherSpec) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	var err error

	m.EncryptedHello, err = bs.ReadBytesRef()
	if err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 返回密码规范切换消息编码后的字节数。
func (m MsgChangeCipherSpec) Size() int {
	return binaryutil.SizeofBytes(m.EncryptedHello)
}

// MsgID 返回密码规范切换消息的内置类型 ID。
func (MsgChangeCipherSpec) MsgID() MsgID {
	return MsgID_ChangeCipherSpec
}

// Clone 深复制加密 Hello 数据。
func (m MsgChangeCipherSpec) Clone() Msg {
	return &MsgChangeCipherSpec{
		EncryptedHello: bytes.Clone(m.EncryptedHello),
	}
}
