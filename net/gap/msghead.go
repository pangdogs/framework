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
	"git.golaxy.org/framework/utils/binaryutil"
	"io"
)

// Origin 来源信息
type Origin struct {
	Svc       string // 服务
	Addr      string // 地址
	Timestamp int64  // 时间戳
}

// Read implements io.Reader
func (o Origin) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteString(o.Svc); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteString(o.Addr); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteInt64(o.Timestamp); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write implements io.Writer
func (o *Origin) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	var err error

	o.Svc, err = bs.ReadString()
	if err != nil {
		return bs.BytesRead(), err
	}

	o.Addr, err = bs.ReadString()
	if err != nil {
		return bs.BytesRead(), err
	}

	o.Timestamp, err = bs.ReadInt64()
	if err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 大小
func (o Origin) Size() int {
	return binaryutil.SizeofString(o.Svc) + binaryutil.SizeofString(o.Addr) + binaryutil.SizeofInt64()
}

// MsgHead 消息头
type MsgHead struct {
	Len   uint32 // 消息长度
	MsgId MsgId  // 消息Id
	Src   Origin // 源信息
	Seq   int64  // 序号
}

// Read implements io.Reader
func (m MsgHead) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteUint32(m.Len); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteUint32(m.MsgId); err != nil {
		return bs.BytesWritten(), err
	}
	if _, err := binaryutil.CopyToByteStream(&bs, m.Src); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteInt64(m.Seq); err != nil {
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

	m.MsgId, err = bs.ReadUint32()
	if err != nil {
		return bs.BytesRead(), err
	}

	_, err = bs.WriteTo(&m.Src)
	if err != nil {
		return bs.BytesRead(), err
	}

	m.Seq, err = bs.ReadInt64()
	if err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 大小
func (m MsgHead) Size() int {
	return binaryutil.SizeofUint32() + binaryutil.SizeofUint32() + m.Src.Size() + binaryutil.SizeofInt64()
}
