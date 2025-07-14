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
	"strings"
)

// Code 错误码
type Code = int32

const (
	Code_VersionError    Code = iota + 1 // 版本错误
	Code_SessionNotFound                 // Session未找到
	Code_EncryptFailed                   // 加密失败
	Code_AuthFailed                      // 鉴权失败
	Code_ContinueFailed                  // 重连失败
	Code_Reject                          // 拒绝连接
	Code_Shutdown                        // 服务关闭
	Code_SessionDeath                    // 会话过期
	Code_Customize       = 32            // 自定义错误码起点
)

// MsgRst 重置链路（注意：为了提高解码性能，减少内存碎片，解码string与bytes字段时均使用引用类型，引用字节池中的bytes，GC时会被归还字节池，不要直接持有此类型字段）
type MsgRst struct {
	Code    Code   // 错误码
	Message string // 错误信息
}

// Read implements io.Reader
func (m *MsgRst) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteInt32(m.Code); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteString(m.Message); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write implements io.Writer
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

// Size 大小
func (m *MsgRst) Size() int {
	return binaryutil.SizeofInt32() + binaryutil.SizeofString(m.Message)
}

// MsgId 消息Id
func (*MsgRst) MsgId() MsgId {
	return MsgId_Rst
}

// Clone 克隆消息对象
func (m *MsgRst) Clone() Msg {
	return &MsgRst{
		Code:    m.Code,
		Message: strings.Clone(m.Message),
	}
}
