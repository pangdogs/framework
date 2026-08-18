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

	"git.golaxy.org/framework/net/gap/variant"
	"git.golaxy.org/framework/utils/binaryutil"
	"git.golaxy.org/framework/utils/correlation"
)

// MsgRPCReply 表示 RPC 请求的响应。
type MsgRPCReply struct {
	CorrID correlation.ID // 对应请求的关联 ID。
	Rets   variant.Array  // 调用返回值。
	Error  variant.Error  // 调用错误；OK 为 true 时表示成功。
}

// Read 将 RPC 响应编码到 p。
func (m MsgRPCReply) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteUvarint(uint64(m.CorrID)); err != nil {
		return bs.BytesWritten(), err
	}
	if _, err := binaryutil.CopyToByteStream(&bs, m.Rets); err != nil {
		return bs.BytesWritten(), err
	}
	if _, err := binaryutil.CopyToByteStream(&bs, m.Error); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write 从 p 解码 RPC 响应。
func (m *MsgRPCReply) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	var err error

	corrID, err := bs.ReadUvarint()
	if err != nil {
		return bs.BytesRead(), err
	}
	m.CorrID = correlation.ID(corrID)

	if _, err = bs.WriteTo(&m.Rets); err != nil {
		return bs.BytesRead(), err
	}

	if _, err = bs.WriteTo(&m.Error); err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 返回 RPC 响应编码后的字节数。
func (m MsgRPCReply) Size() int {
	return binaryutil.SizeofUvarint(uint64(m.CorrID)) + m.Rets.Size() + m.Error.Size()
}

// MsgID 返回 RPC 响应的内置类型 ID。
func (MsgRPCReply) MsgID() MsgID {
	return MsgID_RPC_Reply
}
