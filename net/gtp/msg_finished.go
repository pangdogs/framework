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

package gtp

import (
	"io"

	"git.golaxy.org/framework/utils/binaryutil"
)

const (
	// Flag_EncryptOK 表示服务端确认加密协商成功。
	Flag_EncryptOK Flag = 1 << (iota + Flag_Customize)
	// Flag_AuthOK 表示服务端确认鉴权成功。
	Flag_AuthOK
	// Flag_ContinueOK 表示服务端确认会话续接成功。
	Flag_ContinueOK
)

// MsgFinished 确认握手完成并交换双方当前的传输序号。
type MsgFinished struct {
	SendSeq uint32 // 发送方当前发送序号。
	RecvSeq uint32 // 发送方当前接收序号。
}

// Read 将握手完成消息编码到 p。
func (m MsgFinished) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteUint32(m.SendSeq); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteUint32(m.RecvSeq); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write 从 p 解码握手完成消息。
func (m *MsgFinished) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	var err error

	m.SendSeq, err = bs.ReadUint32()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.RecvSeq, err = bs.ReadUint32()
	if err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 返回握手完成消息的固定编码字节数。
func (MsgFinished) Size() int {
	return binaryutil.SizeofUint32 + binaryutil.SizeofUint32
}

// MsgID 返回握手完成消息的内置类型 ID。
func (MsgFinished) MsgID() MsgID {
	return MsgID_Finished
}

// Clone 返回消息副本。
func (m MsgFinished) Clone() Msg {
	return &m
}
