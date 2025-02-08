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
	ErrEncode = errors.New("gap-encode") // 编码错误
)

var encoder = MakeEncoder()

// DefaultEncoder 默认消息包编码器
func DefaultEncoder() Encoder {
	return encoder
}

// MakeEncoder 创建消息包编码器
func MakeEncoder() Encoder {
	return Encoder{}
}

// Encoder 消息包编码器
type Encoder struct{}

// Encode 编码消息包
func (Encoder) Encode(src gap.Origin, seq int64, msg gap.MsgReader) (ret binaryutil.RecycleBytes, err error) {
	if msg == nil {
		return binaryutil.NilRecycleBytes, fmt.Errorf("%w: %w: msg is nil", ErrEncode, core.ErrArgs)
	}

	mp := gap.MsgPacket{
		Head: gap.MsgHead{
			MsgId: msg.MsgId(),
			Src:   src,
			Seq:   seq,
		},
		Msg: msg,
	}
	mp.Head.Len = uint32(mp.Size())

	mpBuf := binaryutil.MakeRecycleBytes(int(mp.Head.Len))
	defer func() {
		if !mpBuf.Equal(ret) {
			mpBuf.Release()
		}
	}()

	if _, err := binaryutil.CopyToBuff(mpBuf.Data(), mp); err != nil {
		return binaryutil.NilRecycleBytes, fmt.Errorf("%w: write msg failed, %w", ErrEncode, err)
	}

	return mpBuf, nil
}
