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

// MsgRPCRequest 表示需要响应的 RPC 请求。
type MsgRPCRequest struct {
	CorrID    correlation.ID    // 用于匹配响应与 Future 的关联 ID。
	CallChain variant.CallChain // 调用来源链。
	Path      []byte            // 已编码调用路径；解码时引用输入缓冲区。
	Args      variant.Array     // 调用参数。
}

// Read 将 RPC 请求编码到 p。
func (m MsgRPCRequest) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteUvarint(uint64(m.CorrID)); err != nil {
		return bs.BytesWritten(), err
	}
	if _, err := binaryutil.CopyToByteStream(&bs, m.CallChain); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteBytes(m.Path); err != nil {
		return bs.BytesWritten(), err
	}
	if _, err := binaryutil.CopyToByteStream(&bs, m.Args); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write 从 p 解码 RPC 请求。
func (m *MsgRPCRequest) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	var err error

	corrID, err := bs.ReadUvarint()
	if err != nil {
		return bs.BytesRead(), err
	}
	m.CorrID = correlation.ID(corrID)

	if _, err = bs.WriteTo(&m.CallChain); err != nil {
		return bs.BytesRead(), err
	}

	m.Path, err = bs.ReadBytesRef()
	if err != nil {
		return bs.BytesRead(), err
	}

	if _, err = bs.WriteTo(&m.Args); err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 返回 RPC 请求编码后的字节数。
func (m MsgRPCRequest) Size() int {
	return binaryutil.SizeofUvarint(uint64(m.CorrID)) + m.CallChain.Size() + binaryutil.SizeofBytes(m.Path) + m.Args.Size()
}

// MsgID 返回 RPC 请求的内置类型 ID。
func (MsgRPCRequest) MsgID() MsgID {
	return MsgID_RPC_Request
}
