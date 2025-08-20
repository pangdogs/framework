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
	"git.golaxy.org/framework/utils/binaryutil"
	"io"
)

// Flags 所有标志位
type Flags uint8

// Is 判断标志位
func (f Flags) Is(b Flag) bool {
	return f&Flags(b) != 0
}

// Set 设置标志位
func (f *Flags) Set(b Flag, v bool) *Flags {
	if v {
		*f |= Flags(b)
	} else {
		*f &= ^Flags(b)
	}
	return f
}

// Setd 拷贝并设置标志位
func (f Flags) Setd(b Flag, v bool) Flags {
	if v {
		f |= Flags(b)
	} else {
		f &= ^Flags(b)
	}
	return f
}

func Flags_None() Flags {
	return 0
}

// Flag 标志位
type Flag = uint8

// 固定标志位
const (
	Flag_Encrypted  Flag   = 1 << iota // 已加密
	Flag_Signed                        // 已签名
	Flag_Compressed                    // 已压缩
	Flag_Customize  = iota             // 自定义标志位起点
)

// MsgHead 消息头
type MsgHead struct {
	Len   uint32 // 消息包长度
	MsgId MsgId  // 消息Id
	Flags Flags  // 标志位
	Seq   uint32 // 消息序号
	Ack   uint32 // 应答序号
}

// Read implements io.Reader
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

// Write implements io.Writer
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

// Size 大小
func (MsgHead) Size() int {
	return binaryutil.SizeofUint32() + binaryutil.SizeofUint8() + binaryutil.SizeofUint8() +
		binaryutil.SizeofUint32() + binaryutil.SizeofUint32()
}
