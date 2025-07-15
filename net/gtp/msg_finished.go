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
	"git.golaxy.org/framework/utils/binaryutil"
	"io"
)

// Finished消息标志位
const (
	Flag_EncryptOK  Flag = 1 << (iota + Flag_Customize) // 加密成功，在服务端发起的Finished消息携带
	Flag_AuthOK                                         // 鉴权成功，在服务端发起的Finished消息携带
	Flag_ContinueOK                                     // 断线重连成功，在服务端发起的Finished消息携带
)

// MsgFinished 握手结束，表示认可对端，可以开始传输数据
type MsgFinished struct {
	SendSeq uint32 // 服务端请求序号
	RecvSeq uint32 // 服务端响应序号
}

// Read implements io.Reader
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

// Write implements io.Writer
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

// Size 大小
func (MsgFinished) Size() int {
	return binaryutil.SizeofUint32() + binaryutil.SizeofUint32()
}

// MsgId 消息Id
func (MsgFinished) MsgId() MsgId {
	return MsgId_Finished
}

// Clone 克隆消息对象
func (m MsgFinished) Clone() Msg {
	return &m
}
