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

// SyncTime消息标志位
const (
	Flag_ReqTime  Flag = 1 << (iota + Flag_Customize) // 请求同步时间
	Flag_RespTime                                     // 响应同步时间
)

// MsgSyncTime 同步时间
type MsgSyncTime struct {
	CorrId       int64 // 关联Id，用于支持Future等异步模型
	OriginTime   int64 // NTP t1，请求方发送请求时间
	ReceiveTime  int64 // NTP t2，响应方收到请求时间（响应时有效）
	TransmitTime int64 // NTP t3，响应方发送响应时间（响应时有效）
	ZoneOffset   int32 // 响应方时区偏移秒数（响应时有效）
}

// Read implements io.Reader
func (m MsgSyncTime) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteInt64(m.CorrId); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteInt64(m.OriginTime); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteInt64(m.ReceiveTime); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteInt64(m.TransmitTime); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteInt32(m.ZoneOffset); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write implements io.Writer
func (m *MsgSyncTime) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	var err error

	m.CorrId, err = bs.ReadInt64()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.OriginTime, err = bs.ReadInt64()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.ReceiveTime, err = bs.ReadInt64()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.TransmitTime, err = bs.ReadInt64()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.ZoneOffset, err = bs.ReadInt32()
	if err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 大小
func (m MsgSyncTime) Size() int {
	return binaryutil.SizeofInt64 + binaryutil.SizeofInt64 + binaryutil.SizeofInt64 + binaryutil.SizeofInt64 +
		binaryutil.SizeofInt32
}

// MsgId 消息Id
func (MsgSyncTime) MsgId() MsgId {
	return MsgId_SyncTime
}

// Clone 克隆消息对象
func (m MsgSyncTime) Clone() Msg {
	return &m
}
