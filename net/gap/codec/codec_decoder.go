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
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/framework/net/gap"
	"io"
)

var decoder = MakeDecoder(gap.DefaultMsgCreator())

// DefaultDecoder 默认消息包解码器
func DefaultDecoder() Decoder {
	return decoder
}

// MakeDecoder 创建消息包解码器
func MakeDecoder(msgCreator gap.IMsgCreator) Decoder {
	if msgCreator == nil {
		exception.Panicf("gap-dec: %w: msgCreator is nil", core.ErrArgs)
	}
	return Decoder{
		MsgCreator: msgCreator,
	}
}

// Decoder 消息包解码器
type Decoder struct {
	MsgCreator gap.IMsgCreator // 消息对象构建器
}

// Decode 解码消息包
func (d Decoder) Decode(data []byte) (gap.MsgPacket, error) {
	if d.MsgCreator == nil {
		return gap.MsgPacket{}, errors.New("gap-dec: MsgCreator is nil")
	}

	mp := gap.MsgPacket{}

	// 读取消息头
	n, err := mp.Head.Write(data)
	if err != nil {
		return gap.MsgPacket{}, fmt.Errorf("gap-dec: read msg-packet-head failed, %w", err)
	}

	if len(data) < int(mp.Head.Len) {
		return gap.MsgPacket{}, fmt.Errorf("gap-dec: %w (%d < %d)", io.ErrShortBuffer, len(data), mp.Head.Len)
	}

	// 创建消息体
	msg, err := d.MsgCreator.New(mp.Head.MsgId)
	if err != nil {
		return gap.MsgPacket{}, fmt.Errorf("gap-dec: new msg failed, %w (%d)", err, mp.Head.MsgId)
	}

	// 读取消息
	if _, err = msg.Write(data[n:]); err != nil {
		return gap.MsgPacket{}, fmt.Errorf("gap-dec: read msg failed, %w", err)
	}

	mp.Msg = msg

	return mp, nil
}
