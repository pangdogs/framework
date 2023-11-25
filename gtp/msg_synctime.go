package gtp

import (
	"kit.golaxy.org/plugins/util/binaryutil"
)

// SyncTime消息标志位
const (
	Flag_ReqTime  Flag = 1 << (iota + Flag_Customize) // 请求同步时间
	Flag_RespTime                                     // 响应同步时间
)

// MsgSyncTime 同步时间
type MsgSyncTime struct {
	CorrId          int64 // 关联Id，用于支持Future等异步模型
	LocalUnixMilli  int64 // 本地时间
	RemoteUnixMilli int64 // 对端时间（响应时有效）
}

// Read implements io.Reader
func (m MsgSyncTime) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteInt64(m.CorrId); err != nil {
		return 0, err
	}
	if err := bs.WriteInt64(m.LocalUnixMilli); err != nil {
		return 0, err
	}
	if err := bs.WriteInt64(m.RemoteUnixMilli); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (m *MsgSyncTime) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	corrId, err := bs.ReadInt64()
	if err != nil {
		return 0, err
	}
	localUnixMilli, err := bs.ReadInt64()
	if err != nil {
		return 0, err
	}
	remoteUnixMilli, err := bs.ReadInt64()
	if err != nil {
		return 0, err
	}
	m.CorrId = corrId
	m.LocalUnixMilli = localUnixMilli
	m.RemoteUnixMilli = remoteUnixMilli
	return bs.BytesRead(), nil
}

// Size 大小
func (MsgSyncTime) Size() int {
	return binaryutil.SizeofInt64() + binaryutil.SizeofInt64() + binaryutil.SizeofInt64()
}

// MsgId 消息Id
func (MsgSyncTime) MsgId() MsgId {
	return MsgId_SyncTime
}

// Clone 克隆消息对象
func (m MsgSyncTime) Clone() Msg {
	return &m
}
