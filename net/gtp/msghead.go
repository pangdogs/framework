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

// Flags 是消息头中的标志位集合。
type Flags uint8

// Is 报告指定标志位是否已设置。
func (f Flags) Is(b Flag) bool {
	return f&Flags(b) != 0
}

// Set 原地设置或清除指定标志位，并返回接收者。
func (f *Flags) Set(b Flag, v bool) *Flags {
	if v {
		*f |= Flags(b)
	} else {
		*f &= ^Flags(b)
	}
	return f
}

// Setd 在副本上设置或清除指定标志位并返回副本。
func (f Flags) Setd(b Flag, v bool) Flags {
	if v {
		f |= Flags(b)
	} else {
		f &= ^Flags(b)
	}
	return f
}

// Flags_None 返回不包含任何标志位的集合。
func Flags_None() Flags {
	return 0
}

// Flag 表示一个消息头标志位掩码。
type Flag = uint8

const (
	// Flag_Encrypted 表示消息体已加密。
	Flag_Encrypted Flag = 1 << iota
	// Flag_Signed 表示消息体附带认证码。
	Flag_Signed
	// Flag_Compressed 表示消息体已压缩。
	Flag_Compressed
	// Flag_Customize 是自定义标志位的起始位序号。
	Flag_Customize = iota
)

// MsgHead 是每个 GTP 消息包的固定长度头部。
type MsgHead struct {
	Len   uint32 // 完整消息包的字节数。
	MsgId MsgId  // 消息类型 ID。
	Flags Flags  // 加密、认证和压缩标志。
	Seq   uint32 // 当前消息序号。
	Ack   uint32 // 已确认接收的消息序号。
}

// Read 将消息头编码到 p。
func (m MsgHead) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteUint32(m.Len); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteUint8(m.MsgId); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteUint8(uint8(m.Flags)); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteUint32(m.Seq); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteUint32(m.Ack); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write 从 p 解码消息头。
func (m *MsgHead) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	var err error

	m.Len, err = bs.ReadUint32()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.MsgId, err = bs.ReadUint8()
	if err != nil {
		return bs.BytesRead(), err
	}

	flags, err := bs.ReadUint8()
	if err != nil {
		return bs.BytesRead(), err
	}
	m.Flags = Flags(flags)

	m.Seq, err = bs.ReadUint32()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.Ack, err = bs.ReadUint32()
	if err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 返回消息头的固定编码字节数。
func (MsgHead) Size() int {
	return binaryutil.SizeofUint32 + binaryutil.SizeofUint8 + binaryutil.SizeofUint8 +
		binaryutil.SizeofUint32 + binaryutil.SizeofUint32
}
