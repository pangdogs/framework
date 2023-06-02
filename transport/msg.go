package transport

import (
	"kit.golaxy.org/plugins/transport/binaryutil"
)

// MsgId 消息Id
type MsgId = uint8

const (
	MsgId_Hello             MsgId = iota // Hello Handshake C<->S 不加密
	MsgId_SecretKeyExchange              // 秘钥交换 Handshake C<->S 不加密
	MsgId_ChangeCipherSpec               // 变更密码规范 Handshake C<->S 不加密
	MsgId_Auth                           // 鉴权 Handshake C->S 加密
	MsgId_Finished                       // 握手结束 Handshake C<->S 加密
	MsgId_Rst                            // 重置链路 Ctrl S->C 加密
	MsgId_Heartbeat                      // 心跳 Ctrl C<->S 加密
	MsgId_SyncTime                       // 时钟同步 Ctrl S->C 加密
	MsgId_Payload                        // 数据传输 TRANS C<->S 加密
)

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

// MsgHead 消息头
type MsgHead struct {
	Len   uint64 // 消息长度
	Flags Flags  // 标志位
	MsgId MsgId  // 消息Id
}

func (m *MsgHead) Read(p []byte) (int, error) {
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

func (m *MsgHead) Write(p []byte) (int, error) {
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

func (m *MsgHead) Size() int {
	return binaryutil.SizeofUvarint(m.Len) + binaryutil.SizeofUint8() + binaryutil.SizeofUint8()
}
