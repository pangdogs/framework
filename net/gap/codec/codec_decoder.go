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
	"io"

	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/framework/net/gap"
)

var (
	// ErrDecode 是 GAP 消息包解码错误的根错误。
	ErrDecode = errors.New("gap-decode")
)

var decoder = &Decoder{MsgCreator: gap.DefaultMsgCreator()}

// NewDecoder 创建使用指定消息构建器的解码器；构建器不得为 nil。
func NewDecoder(msgCreator gap.IMsgCreator) *Decoder {
	if msgCreator == nil {
		exception.Panicf("%w: %w: msgCreator is nil", ErrDecode, core.ErrArgs)
	}
	if msgCreator == decoder.MsgCreator {
		return decoder
	}
	return &Decoder{
		MsgCreator: msgCreator,
	}
}

// Decoder 根据消息头中的类型 ID 构造并解码消息体。
type Decoder struct {
	MsgCreator gap.IMsgCreator // 用于构造消息体的消息构建器。
}

// Decode 从 data 解码一个消息包；消息字段可能直接引用 data，调用方不得提前复用它。
func (d *Decoder) Decode(data []byte) (gap.MsgPacket, error) {
	if d.MsgCreator == nil {
		return gap.MsgPacket{}, fmt.Errorf("%w: MsgCreator is nil", ErrDecode)
	}

	mp := gap.MsgPacket{}

	// 先解析消息头，以确定消息类型和完整包长。
	n, err := mp.Head.Write(data)
	if err != nil {
		return gap.MsgPacket{}, fmt.Errorf("%w: read msg-packet-head failed, %w", ErrDecode, err)
	}

	if len(data) < int(mp.Head.Len) {
		return gap.MsgPacket{}, fmt.Errorf("%w: %w (%d < %d)", ErrDecode, io.ErrShortBuffer, len(data), mp.Head.Len)
	}

	// 按消息类型构造具体消息；未知类型由 MsgCreator 返回错误。
	msg, err := d.MsgCreator.New(mp.Head.MsgID)
	if err != nil {
		return gap.MsgPacket{}, fmt.Errorf("%w: new msg failed, %w (%d)", ErrDecode, err, mp.Head.MsgID)
	}

	// 消息的 Write 直接接收输入子切片，引用型字段可能与 data 共享底层存储。
	if _, err = msg.Write(data[n:]); err != nil {
		return gap.MsgPacket{}, fmt.Errorf("%w: read msg failed, %w", ErrDecode, err)
	}

	mp.Body = msg

	return mp, nil
}
