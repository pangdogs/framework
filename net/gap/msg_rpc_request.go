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
	"git.golaxy.org/framework/net/gap/variant"
	"git.golaxy.org/framework/utils/binaryutil"
	"io"
)

// MsgRPCRequest RPC请求
type MsgRPCRequest struct {
	CorrId    int64             // 关联Id，用于支持Future等异步模型
	CallChain variant.CallChain // 调用链
	Path      []byte            // 调用路径
	Args      variant.Array     // 参数列表
}

// Read implements io.Reader
func (m MsgRPCRequest) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteVarint(m.CorrId); err != nil {
		return bs.BytesWritten(), err
	}
	if _, err := binaryutil.ReadTo(&bs, m.CallChain); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteBytes(m.Path); err != nil {
		return bs.BytesWritten(), err
	}
	if _, err := binaryutil.ReadTo(&bs, m.Args); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write implements io.Writer
func (m *MsgRPCRequest) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	var err error

	m.CorrId, err = bs.ReadVarint()
	if err != nil {
		return bs.BytesRead(), err
	}

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

// Size 大小
func (m MsgRPCRequest) Size() int {
	return binaryutil.SizeofVarint(m.CorrId) + m.CallChain.Size() + binaryutil.SizeofBytes(m.Path) + m.Args.Size()
}

// MsgId 消息Id
func (MsgRPCRequest) MsgId() MsgId {
	return MsgId_RPC_Request
}
