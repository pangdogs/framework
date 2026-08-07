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
	// Flag_ReqTime 表示时钟同步请求。
	Flag_ReqTime Flag = 1 << (iota + Flag_Customize)
	// Flag_RespTime 表示时钟同步响应。
	Flag_RespTime
)

// MsgSyncTime 携带一次 NTP 风格时钟采样的时间点和时区信息。
type MsgSyncTime struct {
	CorrId       int64 // 请求与响应的关联 ID。
	OriginTime   int64 // NTP t1，请求方发送请求的 Unix 毫秒时间戳。
	ReceiveTime  int64 // NTP t2，响应方收到请求的 Unix 毫秒时间戳。
	TransmitTime int64 // NTP t3，响应方发送响应的 Unix 毫秒时间戳。
	ZoneOffset   int32 // 响应方相对 UTC 的时区偏移秒数。
}

// Read 将时钟同步消息编码到 p。
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

// Write 从 p 解码时钟同步消息。
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

// Size 返回时钟同步消息的固定编码字节数。
func (m MsgSyncTime) Size() int {
	return binaryutil.SizeofInt64 + binaryutil.SizeofInt64 + binaryutil.SizeofInt64 + binaryutil.SizeofInt64 +
		binaryutil.SizeofInt32
}

// MsgId 返回时钟同步消息的内置类型 ID。
func (MsgSyncTime) MsgId() MsgId {
	return MsgId_SyncTime
}

// Clone 返回消息副本。
func (m MsgSyncTime) Clone() Msg {
	return &m
}
