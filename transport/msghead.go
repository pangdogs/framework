package transport

import "kit.golaxy.org/plugins/transport/binaryutil"

// Flags 所有标志位
type Flags uint8

// Is 判断标志位
func (f Flags) Is(b Flag) bool {
	return f&(1<<b) != 0
}

// Set 设置标志位
func (f *Flags) Set(b Flag, v bool) {
	if v {
		*f |= Flags(b)
	} else {
		*f &= ^Flags(b)
	}
}

// Flag 标志位
type Flag = uint8

// 固定标志位
const (
	Flag_Encrypted  Flag   = 1 << iota // 已加密
	Flag_MAC                           // 有MAC
	Flag_Compressed                    // 已压缩
	Flag_Customize  = iota             // 自定义标志位起点
)

// MsgHead 消息头
type MsgHead struct {
	Len   uint32 // 消息长度
	Flags Flags  // 标志位
	MsgId MsgId  // 消息Id
}

func (m *MsgHead) Read(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	if err := bs.WriteUint32(m.Len); err != nil {
		return 0, err
	}
	if err := bs.WriteUint8(m.MsgId); err != nil {
		return 0, err
	}
	if err := bs.WriteUint8(uint8(m.Flags)); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

func (m *MsgHead) Write(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	l, err := bs.ReadUint32()
	if err != nil {
		return 0, err
	}
	msgid, err := bs.ReadUint8()
	if err != nil {
		return 0, err
	}
	flags, err := bs.ReadUint8()
	if err != nil {
		return 0, err
	}
	m.Len = l
	m.MsgId = msgid
	m.Flags = Flags(flags)
	return bs.BytesRead(), nil
}

func (m *MsgHead) Size() int {
	return binaryutil.SizeofUint32() + binaryutil.SizeofUint8() + binaryutil.SizeofUint8()
}
