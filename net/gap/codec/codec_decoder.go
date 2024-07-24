package codec

import (
	"errors"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/framework/net/gap"
	"io"
)

var decoder = MakeDecoder(gap.DefaultMsgCreator())

// DefaultDecoder 默认消息包解码器
func DefaultDecoder() Decoder {
	return decoder
}

// MakeDecoder 创建消息包解码器
func MakeDecoder(mc gap.IMsgCreator) Decoder {
	if mc == nil {
		panic(fmt.Errorf("gap: %w: mc is nil", core.ErrArgs))
	}
	return Decoder{
		MsgCreator: mc,
	}
}

// Decoder 消息包解码器
type Decoder struct {
	MsgCreator gap.IMsgCreator // 消息对象构建器
}

// Decode 解码消息包
func (d Decoder) Decode(data []byte) (gap.MsgPacket, error) {
	if d.MsgCreator == nil {
		return gap.MsgPacket{}, errors.New("gap: setting MsgCreator is nil")
	}

	mp := gap.MsgPacket{}

	// 读取消息头
	n, err := mp.Head.Write(data)
	if err != nil {
		return gap.MsgPacket{}, fmt.Errorf("gap: read msg-packet-head failed, %w", err)
	}

	if len(data) < int(mp.Head.Len) {
		return gap.MsgPacket{}, fmt.Errorf("gap: %w (%d < %d)", io.ErrShortBuffer, len(data), mp.Head.Len)
	}

	// 创建消息体
	msg, err := d.MsgCreator.New(mp.Head.MsgId)
	if err != nil {
		return gap.MsgPacket{}, fmt.Errorf("gap: new msg failed, %w (%d)", err, mp.Head.MsgId)
	}

	// 读取消息
	if _, err = msg.Write(data[n:]); err != nil {
		return gap.MsgPacket{}, fmt.Errorf("gap: read msg failed, %w", err)
	}

	mp.Msg = msg

	return mp, nil
}
