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
	"strings"

	"git.golaxy.org/framework/utils/binaryutil"
)

// Code 标识 GTP 链路重置原因。
type Code = int32

const (
	// Code_VersionError 表示协议版本不兼容。
	Code_VersionError Code = iota + 1
	// Code_SessionNotFound 表示待续接会话不存在。
	Code_SessionNotFound
	// Code_EncryptFailed 表示加密协商或验证失败。
	Code_EncryptFailed
	// Code_AuthFailed 表示鉴权失败。
	Code_AuthFailed
	// Code_ContinueFailed 表示会话续接失败。
	Code_ContinueFailed
	// Code_Reject 表示对端拒绝连接。
	Code_Reject
	// Code_Shutdown 表示服务正在关闭。
	Code_Shutdown
	// Code_SessionDeath 表示会话已过期。
	Code_SessionDeath
	// Code_Customize 是自定义重置错误码的起始值。
	Code_Customize = 32
)

// MsgRst 请求终止链路并说明原因。直接通过 Write 或 Unmarshal 解码时，Message 会引用输入切片；
// 输入将被复用或修改时应先 Clone。Decoder.Decode 返回的消息不引用调用方输入。
type MsgRst struct {
	Code    Code   // 重置原因码。
	Message string // 可读错误消息。
}

// Read 将链路重置消息编码到 p。
func (m MsgRst) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteInt32(m.Code); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteString(m.Message); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write 从 p 解码链路重置消息，Message 会引用 p。
func (m *MsgRst) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	var err error

	code, err := bs.ReadInt32()
	if err != nil {
		return bs.BytesRead(), err
	}
	m.Code = code

	m.Message, err = bs.ReadStringRef()
	if err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 返回链路重置消息编码后的字节数。
func (m MsgRst) Size() int {
	return binaryutil.SizeofInt32 + binaryutil.SizeofString(m.Message)
}

// MsgId 返回链路重置消息的内置类型 ID。
func (MsgRst) MsgId() MsgId {
	return MsgId_Rst
}

// Clone 深复制可读错误消息。
func (m MsgRst) Clone() Msg {
	return &MsgRst{
		Code:    m.Code,
		Message: strings.Clone(m.Message),
	}
}
