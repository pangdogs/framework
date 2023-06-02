package transport

import "kit.golaxy.org/plugins/transport/binaryutil"

// Flag 标志位
type Flag = uint8

// 固定标志位
const (
	Flag_Encrypt  Flag   = 1 << iota // 是否加密
	Flag_Compress                    // 是否压缩
	Flag_MAC                         // 是否有完整性校验码
	Flag_Options  = iota             // 可选标志位起点
)

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

// MsgHeadHandshake Handshake类消息头
type MsgHeadHandshake struct {
	Len   uint64 // 消息长度
	Flags Flags  // 标志位
	MsgId MsgId  // 消息Id
}

func (m *MsgHeadHandshake) Read(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	l, err := bs.ReadUvarint()
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

func (m *MsgHeadHandshake) Write(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	if err := bs.WriteUvarint(m.Len); err != nil {
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

func (m *MsgHeadHandshake) Size() int {
	return binaryutil.SizeofUvarint(m.Len) + binaryutil.SizeofUint8() + binaryutil.SizeofUint8()
}
