package codec

import (
	"bytes"
	"errors"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/framework/plugins/gtp"
	"git.golaxy.org/framework/plugins/util/binaryutil"
	"io"
)

var (
	ErrDataNotEnough = io.ErrShortBuffer // 数据不足
)

// IValidate 验证消息包接口
type IValidate interface {
	// Validate 验证消息包
	Validate(msgHead gtp.MsgHead, msgBuff []byte) error
}

// IDecoder 消息包解码器接口
type IDecoder interface {
	io.Writer
	io.ReaderFrom
	// Reset 重置缓存
	Reset()
	// Decode 从缓存，解码消息包
	Decode(validate ...IValidate) (gtp.MsgPacket, error)
	// DecodeBuff 从指定buff，解码消息包
	DecodeBuff(buff *bytes.Buffer, validate ...IValidate) (gtp.MsgPacket, error)
	// DecodeBytes 从指定bytes，解码消息包
	DecodeBytes(data []byte, validate ...IValidate) (gtp.MsgPacket, error)
	// GC GC
	GC()
}

// Decoder 消息包解码器
type Decoder struct {
	MsgCreator        gtp.IMsgCreator           // 消息对象构建器
	EncryptionModule  IEncryptionModule         // 加密模块
	MACModule         IMACModule                // MAC模块
	CompressionModule ICompressionModule        // 压缩模块
	buffer            bytes.Buffer              // buffer
	gcList            []binaryutil.RecycleBytes // GC列表
}

// Write implements io.Writer
func (d *Decoder) Write(p []byte) (int, error) {
	return d.buffer.Write(p)
}

// ReadFrom implements io.ReaderFrom
func (d *Decoder) ReadFrom(r io.Reader) (int64, error) {
	if r == nil {
		return 0, fmt.Errorf("gtp: %w: r is nil", core.ErrArgs)
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
func (d *Decoder) Decode(validate ...IValidate) (gtp.MsgPacket, error) {
	return d.DecodeBuff(&d.buffer, validate...)
}

// DecodeBuff 从指定buff，解码消息包
func (d *Decoder) DecodeBuff(buff *bytes.Buffer, validate ...IValidate) (gtp.MsgPacket, error) {
	if buff == nil {
		return gtp.MsgPacket{}, fmt.Errorf("gtp: %w: buff is nil", core.ErrArgs)
	}

	// 探测消息包长度
	length, err := d.lengthDetection(buff.Bytes())
	if err != nil {
		return gtp.MsgPacket{}, err
	}

	// 解码后，丢弃消息包数据
	defer buff.Next(length)

	// 解码消息包
	return d.decode(buff.Bytes()[:length], validate...)
}

// DecodeBytes 从指定bytes，解码消息包
func (d *Decoder) DecodeBytes(data []byte, validate ...IValidate) (gtp.MsgPacket, error) {
	// 探测消息包长度
	length, err := d.lengthDetection(data)
	if err != nil {
		return gtp.MsgPacket{}, err
	}

	// 解码消息包
	return d.decode(data[:length], validate...)
}

// GC GC
func (d *Decoder) GC() {
	for i := range d.gcList {
		d.gcList[i].Release()
	}
	d.gcList = d.gcList[:0]
}

// lengthDetection 消息包长度探针
func (d *Decoder) lengthDetection(data []byte) (int, error) {
	mpl := gtp.MsgPacketLen{}

	// 读取消息包长度
	_, err := mpl.Write(data)
	if err != nil {
		return 0, ErrDataNotEnough
	}

	if len(data) < int(mpl.Len) {
		return int(mpl.Len), fmt.Errorf("gtp: %w (%d < %d)", ErrDataNotEnough, len(data), mpl.Len)
	}

	return int(mpl.Len), nil
}

// decode 解码消息包
func (d *Decoder) decode(data []byte, validate ...IValidate) (gtp.MsgPacket, error) {
	if d.MsgCreator == nil {
		return gtp.MsgPacket{}, errors.New("gtp: setting MsgCreator is nil")
	}

	// 消息包数据缓存
	mpBuf := binaryutil.MakeRecycleBytes(binaryutil.BytesPool.Get(len(data)))
	d.gcList = append(d.gcList, mpBuf)

	// 拷贝消息包
	copy(mpBuf.Data(), data)

	mp := gtp.MsgPacket{}

	// 读取消息头
	_, err := mp.Write(mpBuf.Data())
	if err != nil {
		return gtp.MsgPacket{}, fmt.Errorf("gtp: read msg-packet-head failed, %w", err)
	}

	msgBuf := mpBuf.Data()[mp.Head.Size():]

	// 验证消息包
	if len(validate) > 0 {
		err = validate[0].Validate(mp.Head, msgBuf)
		if err != nil {
			return gtp.MsgPacket{}, fmt.Errorf("gtp: validate msg-packet-head failed, %w", err)
		}
	}

	// 检查加密标记
	if mp.Head.Flags.Is(gtp.Flag_Encrypted) {
		// 解密消息体
		if d.EncryptionModule == nil {
			return gtp.MsgPacket{}, errors.New("gtp: setting EncryptionModule is nil, msg can't be decrypted")
		}
		dencryptBuf, err := d.EncryptionModule.Transforming(msgBuf, msgBuf)
		if err != nil {
			return gtp.MsgPacket{}, fmt.Errorf("dencrypt msg failed, %w", err)
		}
		if dencryptBuf.Recyclable() {
			d.gcList = append(d.gcList, dencryptBuf)
		}
		msgBuf = dencryptBuf.Data()

		// 检查MAC标记
		if mp.Head.Flags.Is(gtp.Flag_MAC) {
			if d.MACModule == nil {
				return gtp.MsgPacket{}, errors.New("gtp: setting MACModule is nil, msg can't be verify MAC")
			}
			// 检测MAC
			msgBuf, err = d.MACModule.VerifyMAC(mp.Head.MsgId, mp.Head.Flags, msgBuf)
			if err != nil {
				return gtp.MsgPacket{}, fmt.Errorf("gtp: verify msg-mac failed, %w", err)
			}
		}
	}

	// 检查压缩标记
	if mp.Head.Flags.Is(gtp.Flag_Compressed) {
		if d.CompressionModule == nil {
			return gtp.MsgPacket{}, errors.New("gtp: setting CompressionModule is nil, msg can't be uncompress")
		}
		uncompressedBuf, err := d.CompressionModule.Uncompress(msgBuf)
		if err != nil {
			return gtp.MsgPacket{}, fmt.Errorf("gtp: uncompress msg failed, %w", err)
		}
		if uncompressedBuf.Recyclable() {
			d.gcList = append(d.gcList, uncompressedBuf)
		}
		msgBuf = uncompressedBuf.Data()
	}

	// 创建消息体
	mp.Msg, err = d.MsgCreator.New(mp.Head.MsgId)
	if err != nil {
		return gtp.MsgPacket{}, fmt.Errorf("gtp: new msg failed, %w (%d)", err, mp.Head.MsgId)
	}

	// 读取消息
	_, err = mp.Msg.Write(msgBuf)
	if err != nil {
		return gtp.MsgPacket{}, fmt.Errorf("gtp: read msg failed, %w", err)
	}

	return mp, nil
}
