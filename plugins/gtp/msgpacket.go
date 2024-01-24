package gtp

import (
	"git.golaxy.org/framework/plugins/util/binaryutil"
)

// MsgPacket 消息包
type MsgPacket struct {
	Head MsgHead // 消息头
	Msg  Msg     // 消息
}

// Read implements io.Reader
func (mp MsgPacket) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	if _, err := binaryutil.ReadFrom(&bs, mp.Head); err != nil {
		return bs.BytesWritten(), err
	}

	if mp.Msg == nil {
		return bs.BytesWritten(), nil
	}

	if _, err := binaryutil.ReadFrom(&bs, mp.Msg); err != nil {
		return bs.BytesWritten(), err
	}

	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (mp *MsgPacket) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	if _, err := bs.WriteTo(&mp.Head); err != nil {
		return bs.BytesRead(), err
	}

	if mp.Msg == nil {
		return bs.BytesRead(), nil
	}

	if _, err := bs.WriteTo(mp.Msg); err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 大小
func (mp MsgPacket) Size() int {
	n := mp.Head.Size()

	if mp.Msg != nil {
		n += mp.Msg.Size()
	}

	return n
}

// MsgPacketLen 消息包长度
type MsgPacketLen struct {
	Len uint32 // 消息包长度
}

// Read implements io.Reader
func (m MsgPacketLen) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteUint32(m.Len); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (m *MsgPacketLen) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	var err error

	m.Len, err = bs.ReadUint32()
	if err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 大小
func (MsgPacketLen) Size() int {
	return binaryutil.SizeofUint32()
}
