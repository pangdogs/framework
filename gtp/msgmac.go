package gtp

import (
	"kit.golaxy.org/plugins/util/binaryutil"
)

// MsgMAC32 包含MAC(32bit)消息
type MsgMAC32 struct {
	Data []byte
	MAC  uint32
}

// Read implements io.Reader
func (m MsgMAC32) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteBytes(m.Data); err != nil {
		return 0, err
	}
	if err := bs.WriteUint32(m.MAC); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (m *MsgMAC32) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	data, err := bs.ReadBytesRef()
	if err != nil {
		return 0, err
	}
	mac, err := bs.ReadUint32()
	if err != nil {
		return 0, err
	}
	m.Data = data
	m.MAC = mac
	return bs.BytesRead(), nil
}

// Size 大小
func (m MsgMAC32) Size() int {
	return binaryutil.SizeofBytes(m.Data) + binaryutil.SizeofUint32()
}

// MsgMAC64 包含MAC(64bit)消息
type MsgMAC64 struct {
	Data []byte
	MAC  uint64
}

// Read implements io.Reader
func (m MsgMAC64) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteBytes(m.Data); err != nil {
		return 0, err
	}
	if err := bs.WriteUint64(m.MAC); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (m *MsgMAC64) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	data, err := bs.ReadBytesRef()
	if err != nil {
		return 0, err
	}
	mac, err := bs.ReadUint64()
	if err != nil {
		return 0, err
	}
	m.Data = data
	m.MAC = mac
	return bs.BytesRead(), nil
}

// Size 大小
func (m MsgMAC64) Size() int {
	return binaryutil.SizeofBytes(m.Data) + binaryutil.SizeofUint64()
}

// MsgMAC 包含MAC消息
type MsgMAC struct {
	Data []byte
	MAC  []byte
}

// Read implements io.Reader
func (m MsgMAC) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteBytes(m.Data); err != nil {
		return 0, err
	}
	if err := bs.WriteBytes(m.MAC); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (m *MsgMAC) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	data, err := bs.ReadBytesRef()
	if err != nil {
		return 0, err
	}
	mac, err := bs.ReadBytesRef()
	if err != nil {
		return 0, err
	}
	m.Data = data
	m.MAC = mac
	return bs.BytesRead(), nil
}

// Size 大小
func (m MsgMAC) Size() int {
	return binaryutil.SizeofBytes(m.Data) + binaryutil.SizeofBytes(m.MAC)
}
