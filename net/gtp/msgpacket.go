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

// MsgPacket 消息包
type MsgPacket struct {
	Head MsgHead     // 消息头
	Body ReadableMsg // 消息体
}

// Read implements io.Reader
func (mp MsgPacket) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	if _, err := binaryutil.CopyToByteStream(&bs, mp.Head); err != nil {
		return bs.BytesWritten(), err
	}

	if mp.Body == nil {
		return bs.BytesWritten(), io.EOF
	}

	if _, err := binaryutil.CopyToByteStream(&bs, mp.Body); err != nil {
		return bs.BytesWritten(), err
	}

	return bs.BytesWritten(), io.EOF
}

// Size 大小
func (mp MsgPacket) Size() int {
	n := mp.Head.Size()

	if mp.Body != nil {
		n += mp.Body.Size()
	}

	return n
}

// MsgPacketLen 消息包长度
type MsgPacketLen struct {
	Len uint32 // 消息包长度
}

// Read implements io.Reader
func (m MsgPacketLen) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteUint32(m.Len); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write implements io.Writer
func (m *MsgPacketLen) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	var err error

	m.Len, err = bs.ReadUint32()
	if err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 大小
func (MsgPacketLen) Size() int {
	return binaryutil.SizeofUint32
}
