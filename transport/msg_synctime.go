package transport

import (
	"kit.golaxy.org/plugins/transport/binaryutil"
)

// MsgSyncTime 同步时间
type MsgSyncTime struct {
	UnixMilli int64 // Unix时间（毫秒）
}

// Read implements io.Reader
func (m *MsgSyncTime) Read(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	if err := bs.WriteInt64(m.UnixMilli); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (m *MsgSyncTime) Write(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	unixMilli, err := bs.ReadInt64()
	if err != nil {
		return 0, err
	}
	m.UnixMilli = unixMilli
	return bs.BytesRead(), nil
}

// Size 消息大小
func (m *MsgSyncTime) Size() int {
	return binaryutil.SizeofInt64()
}

// MsgId 消息Id
func (MsgSyncTime) MsgId() MsgId {
	return MsgId_SyncTime
}

// Clone 克隆消息对象
func (m MsgSyncTime) Clone() Msg {
	return &m
}
