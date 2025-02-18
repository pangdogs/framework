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
	"git.golaxy.org/framework/utils/binaryutil"
	"io"
	"time"
)

type Call struct {
	Svc       string    // 服务
	Addr      string    // 地址
	Timestamp time.Time // 时间戳
	Transit   bool      // 是否为中转
}

type CallChain []Call

// Read implements io.Reader
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

// Write implements io.Writer
func (v *CallChain) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	l, err := bs.ReadUvarint()
	if err != nil {
		return bs.BytesRead(), err
	}

	*v = make([]Call, l)

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

		(*v)[i].Svc = svc
		(*v)[i].Addr = addr
		(*v)[i].Transit = transit
		(*v)[i].Timestamp = time.UnixMilli(ts).Local()
	}

	return bs.BytesRead(), nil
}

// Size 大小
func (v CallChain) Size() int {
	n := binaryutil.SizeofUvarint(uint64(len(v)))
	for i := range v {
		n += binaryutil.SizeofString(v[i].Svc)
		n += binaryutil.SizeofString(v[i].Addr)
		n += binaryutil.SizeofInt64()
		n += binaryutil.SizeofBool()
	}
	return n
}

// TypeId 类型
func (CallChain) TypeId() TypeId {
	return TypeId_CallChain
}

// Indirect 原始值
func (v CallChain) Indirect() any {
	return v
}

func (v CallChain) First() Call {
	if len(v) <= 0 {
		return Call{}
	}
	return v[0]
}

func (v CallChain) Last() Call {
	if len(v) <= 0 {
		return Call{}
	}
	return v[len(v)-1]
}
