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
)

// MsgOneWayRPC 单程RPC请求
type MsgOneWayRPC struct {
	CallChain variant.CallChain // 调用链
	Path      string            // 调用路径
	Args      variant.Array     // 参数列表
}

// Read implements io.Reader
func (m MsgOneWayRPC) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if _, err := binaryutil.ReadFrom(&bs, m.CallChain); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteString(m.Path); err != nil {
		return bs.BytesWritten(), err
	}
	if _, err := binaryutil.ReadFrom(&bs, m.Args); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (m *MsgOneWayRPC) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	var err error

	if _, err = bs.WriteTo(&m.CallChain); err != nil {
		return bs.BytesRead(), err
	}

	m.Path, err = bs.ReadString()
	if err != nil {
		return bs.BytesRead(), err
	}

	if _, err = bs.WriteTo(&m.Args); err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 大小
func (m MsgOneWayRPC) Size() int {
	return m.CallChain.Size() + binaryutil.SizeofString(m.Path) + m.Args.Size()
}

// MsgId 消息Id
func (MsgOneWayRPC) MsgId() MsgId {
	return MsgId_OneWayRPC
}
