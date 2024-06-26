package codec

import (
	"bytes"
	"errors"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/framework/net/gap"
	"io"
)

var (
	ErrDataNotEnough = io.ErrShortBuffer // 数据不足
)

// DefaultDecoder 默认消息包解码器
func DefaultDecoder() Decoder {
	return MakeDecoder(gap.DefaultMsgCreator())
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
	buffer     bytes.Buffer    // buffer
}

// Write implements io.Writer
func (d *Decoder) Write(p []byte) (int, error) {
	return d.buffer.Write(p)
}

// ReadFrom implements io.ReaderFrom
func (d *Decoder) ReadFrom(r io.Reader) (int64, error) {
	if r == nil {
		return 0, fmt.Errorf("gap: %w: r is nil", core.ErrArgs)
	}

	var buff [bytes.MinRead]byte

	n, err := r.Read(buff[:])
	if n > 0 {
		d.buffer.Write(buff[:n])
	}

	return int64(n), err
}

// Reset 重置缓存
func (d *Decoder) Reset() {
	d.buffer.Reset()
}

// Decode 从缓存，解码消息包
func (d *Decoder) Decode() (gap.MsgPacket, error) {
	return d.DecodeBuff(&d.buffer)
}

// DecodeBuff 从指定buff，解码消息包
func (d Decoder) DecodeBuff(buff *bytes.Buffer) (gap.MsgPacket, error) {
	mp, n, err := d.decode(buff.Bytes())
	buff.Next(n)
	return mp, err
}

// DecodeBytes 从指定bytes，解码消息包
func (d Decoder) DecodeBytes(data []byte) (gap.MsgPacket, error) {
	mp, _, err := d.decode(data)
	return mp, err
}

// decode 解码消息包
func (d Decoder) decode(data []byte) (gap.MsgPacket, int, error) {
	if d.MsgCreator == nil {
		return gap.MsgPacket{}, 0, errors.New("gap: setting MsgCreator is nil")
	}

	mp := gap.MsgPacket{}

	// 读取消息头
	n, err := mp.Head.Write(data)
	if err != nil {
		return gap.MsgPacket{}, 0, fmt.Errorf("gap: read msg-packet-head failed, %w", err)
	}

	if len(data) < int(mp.Head.Len) {
		return gap.MsgPacket{}, 0, fmt.Errorf("gap: %w (%d < %d)", ErrDataNotEnough, len(data), mp.Head.Len)
	}

	// 创建消息体
	msg, err := d.MsgCreator.New(mp.Head.MsgId)
	if err != nil {
		return gap.MsgPacket{}, int(mp.Head.Len), fmt.Errorf("gap: new msg failed, %w (%d)", err, mp.Head.MsgId)
	}

	// 读取消息
	_, err = msg.Write(data[n:])
	if err != nil {
		return gap.MsgPacket{}, int(mp.Head.Len), fmt.Errorf("gap: read msg failed, %w", err)
	}

	mp.Msg = msg

	return mp, int(mp.Head.Len), nil
}
