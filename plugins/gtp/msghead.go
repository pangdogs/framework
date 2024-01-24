package gtp

import (
	"git.golaxy.org/framework/plugins/util/binaryutil"
)

// Flags 所有标志位
type Flags uint8

// Is 判断标志位
func (f Flags) Is(b Flag) bool {
	return f&Flags(b) != 0
}

// Set 设置标志位
func (f *Flags) Set(b Flag, v bool) *Flags {
	if v {
		*f |= Flags(b)
	} else {
		*f &= ^Flags(b)
	}
	return f
}

// Setd 拷贝并设置标志位
func (f Flags) Setd(b Flag, v bool) Flags {
	if v {
		f |= Flags(b)
	} else {
		f &= ^Flags(b)
	}
	return f
}

func Flags_None() Flags {
	return 0
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
	Len   uint32 // 消息包长度
	MsgId MsgId  // 消息Id
	Flags Flags  // 标志位
	Seq   uint32 // 消息序号
	Ack   uint32 // 应答序号
}

// Read implements io.Reader
func (m MsgHead) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteUint32(m.Len); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteUint8(m.MsgId); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteUint8(uint8(m.Flags)); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteUint32(m.Seq); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteUint32(m.Ack); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (m *MsgHead) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	var err error

	m.Len, err = bs.ReadUint32()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.MsgId, err = bs.ReadUint8()
	if err != nil {
		return bs.BytesRead(), err
	}

	flags, err := bs.ReadUint8()
	if err != nil {
		return bs.BytesRead(), err
	}
	m.Flags = Flags(flags)

	m.Seq, err = bs.ReadUint32()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.Ack, err = bs.ReadUint32()
	if err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 大小
func (MsgHead) Size() int {
	return binaryutil.SizeofUint32() + binaryutil.SizeofUint8() + binaryutil.SizeofUint8() +
		binaryutil.SizeofUint32() + binaryutil.SizeofUint32()
}
