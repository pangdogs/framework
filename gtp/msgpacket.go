package gtp

import (
	"kit.golaxy.org/plugins/gtp/binaryutil"
)

// MsgPacket 消息包
type MsgPacket struct {
	Head MsgHead // 消息头
	Msg  Msg     // 消息
}

// Read implements io.Reader
func (mp *MsgPacket) Read(p []byte) (int, error) {
	hn, err := mp.Head.Read(p)
	if err != nil {
		return hn, err
	}
	if mp.Msg == nil {
		return hn, nil
	}
	mn, err := mp.Msg.Read(p[hn:])
	return mn + hn, err
}

// Write implements io.Writer
func (mp *MsgPacket) Write(p []byte) (int, error) {
	hn, err := mp.Head.Write(p)
	if err != nil {
		return hn, err
	}
	if mp.Msg == nil {
		return hn, nil
	}
	mn, err := mp.Msg.Write(p[hn:])
	return mn + hn, err
}

// Size 大小
func (mp *MsgPacket) Size() int {
	if mp.Msg == nil {
		return mp.Head.Size()
	}
	return mp.Head.Size() + mp.Msg.Size()
}

// MsgPacketLen 消息包长度
type MsgPacketLen struct {
	Len uint32 // 消息包长度
}

// Read implements io.Reader
func (m *MsgPacketLen) Read(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	if err := bs.WriteUint32(m.Len); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (m *MsgPacketLen) Write(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	l, err := bs.ReadUint32()
	if err != nil {
		return 0, err
	}
	m.Len = l
	return bs.BytesRead(), nil
}

// Size 大小
func (MsgPacketLen) Size() int {
	return binaryutil.SizeofUint32()
}
