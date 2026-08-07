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

package variant

import (
	"io"
	"time"

	"git.golaxy.org/framework/utils/binaryutil"
)

// Call 描述 RPC 调用链中的一个来源或中转节点。
type Call struct {
	Svc       string    // 服务名。
	Addr      string    // 通信地址。
	Timestamp time.Time // 产生或转发调用的时间。
	Transit   bool      // 是否为中转节点。
}

// CallChain 按调用传播顺序保存来源和中转节点。
type CallChain []Call

// Read 将调用链编码到 p。
func (v CallChain) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	if err := bs.WriteUvarint(uint64(len(v))); err != nil {
		return bs.BytesWritten(), err
	}

	for i := range v {
		if err := bs.WriteString(v[i].Svc); err != nil {
			return bs.BytesWritten(), err
		}
		if err := bs.WriteString(v[i].Addr); err != nil {
			return bs.BytesWritten(), err
		}
		if err := bs.WriteInt64(v[i].Timestamp.UnixMilli()); err != nil {
			return bs.BytesWritten(), err
		}
		if err := bs.WriteBool(v[i].Transit); err != nil {
			return bs.BytesWritten(), err
		}
	}

	return bs.BytesWritten(), io.EOF
}

// Write 从 p 解码调用链，时间戳转换为本地时区。
func (v *CallChain) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	l, err := bs.ReadUvarint()
	if err != nil {
		return bs.BytesRead(), err
	}

	*v = make([]Call, 0, min(l, 256))

	for i := uint64(0); i < l; i++ {
		svc, err := bs.ReadString()
		if err != nil {
			return bs.BytesRead(), err
		}
		addr, err := bs.ReadString()
		if err != nil {
			return bs.BytesRead(), err
		}
		ts, err := bs.ReadInt64()
		if err != nil {
			return bs.BytesRead(), err
		}
		transit, err := bs.ReadBool()
		if err != nil {
			return bs.BytesRead(), err
		}
		item := Call{
			Svc:       svc,
			Addr:      addr,
			Timestamp: time.UnixMilli(ts).Local(),
			Transit:   transit,
		}
		*v = append(*v, item)
	}

	return bs.BytesRead(), nil
}

// Size 返回调用链编码后的字节数。
func (v CallChain) Size() int {
	n := binaryutil.SizeofUvarint(uint64(len(v)))
	for i := range v {
		n += binaryutil.SizeofString(v[i].Svc)
		n += binaryutil.SizeofString(v[i].Addr)
		n += binaryutil.SizeofInt64
		n += binaryutil.SizeofBool
	}
	return n
}

// TypeId 返回调用链的内置类型 ID。
func (CallChain) TypeId() TypeId {
	return TypeId_CallChain
}

// Indirect 返回调用链值本身。
func (v CallChain) Indirect() any {
	return v
}

// First 返回调用链首项；空调用链返回零值。
func (v CallChain) First() Call {
	if len(v) <= 0 {
		return Call{}
	}
	return v[0]
}

// Last 返回调用链末项；空调用链返回零值。
func (v CallChain) Last() Call {
	if len(v) <= 0 {
		return Call{}
	}
	return v[len(v)-1]
}
