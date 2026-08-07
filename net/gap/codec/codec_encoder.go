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

package codec

import (
	"errors"
	"fmt"

	"git.golaxy.org/core"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/utils/binaryutil"
)

var (
	// ErrEncode 是 GAP 消息包编码错误的根错误。
	ErrEncode = errors.New("gap-encode")
)

var encoder = &Encoder{}

// NewEncoder 返回无状态的共享消息包编码器。
func NewEncoder() *Encoder {
	return encoder
}

// Encoder 将 GAP 消息和来源信息编码为完整消息包。
type Encoder struct{}

// Encode 编码消息包并返回池化字节缓冲区；调用方使用完后必须调用 Release。
func (*Encoder) Encode(src gap.Origin, seq int64, msg gap.ReadableMsg) (ret binaryutil.Bytes, err error) {
	if msg == nil {
		return binaryutil.EmptyBytes, fmt.Errorf("%w: %w: msg is nil", ErrEncode, core.ErrArgs)
	}

	mp := gap.MsgPacket{
		Head: gap.MsgHead{
			MsgId: msg.MsgId(),
			Src:   src,
			Seq:   seq,
		},
		Body: msg,
	}
	mp.Head.Len = uint32(mp.Size())

	mpBuf := binaryutil.NewBytes(true, int(mp.Head.Len))
	defer func() {
		if !mpBuf.SameRef(ret) {
			mpBuf.Release()
		}
	}()

	if _, err := binaryutil.CopyToBuff(mpBuf.Payload(), mp); err != nil {
		return binaryutil.EmptyBytes, fmt.Errorf("%w: write msg failed, %w", ErrEncode, err)
	}

	return mpBuf, nil
}
