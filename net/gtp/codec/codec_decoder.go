package codec

import (
	"errors"
	"fmt"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/utils/binaryutil"
	"io"
)

// IValidate 验证消息包接口
type IValidate interface {
	// Validate 验证消息包
	Validate(msgHead gtp.MsgHead, msgBuf []byte) error
}

// IDecoder 消息包解码器接口
type IDecoder interface {
	// Decode 解码消息包
	Decode(data []byte, validate ...IValidate) (gtp.MsgPacket, int, error)
	// GC GC
	GC()
}

// Decoder 消息包解码器
type Decoder struct {
	MsgCreator        gtp.IMsgCreator           // 消息对象构建器
	EncryptionModule  IEncryptionModule         // 加密模块
	MACModule         IMACModule                // MAC模块
	CompressionModule ICompressionModule        // 压缩模块
	gcList            []binaryutil.RecycleBytes // GC列表
}

// Decode 解码消息包
func (d *Decoder) Decode(data []byte, validate ...IValidate) (gtp.MsgPacket, int, error) {
	// 探测消息包长度
	length, err := d.lengthDetection(data)
	if err != nil {
		return gtp.MsgPacket{}, length, err
	}

	// 解码消息包
	mp, err := d.decode(data[:length], validate...)
	if err != nil {
		return gtp.MsgPacket{}, length, err
	}

	return mp, length, nil
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
	if _, err := mpl.Write(data); err != nil {
		return 0, io.ErrShortBuffer
	}

	if len(data) < int(mpl.Len) {
		return int(mpl.Len), fmt.Errorf("gtp: %w (%d < %d)", io.ErrShortBuffer, len(data), mpl.Len)
	}

	return int(mpl.Len), nil
}

// decode 解码消息包
func (d *Decoder) decode(data []byte, validate ...IValidate) (gtp.MsgPacket, error) {
	if d.MsgCreator == nil {
		return gtp.MsgPacket{}, errors.New("gtp: setting MsgCreator is nil")
	}

	// 消息包数据缓存
	mpBuf := binaryutil.MakeRecycleBytes(len(data))
	d.gcList = append(d.gcList, mpBuf)

	// 拷贝消息包
	copy(mpBuf.Data(), data)

	mp := gtp.MsgPacket{}

	// 读取消息头
	if _, err := mp.Head.Write(mpBuf.Data()); err != nil {
		return gtp.MsgPacket{}, fmt.Errorf("gtp: read msg-packet-head failed, %w", err)
	}

	msgBuf := mpBuf.Data()[mp.Head.Size():]

	// 验证消息包
	if len(validate) > 0 {
		if err := validate[0].Validate(mp.Head, msgBuf); err != nil {
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
	msg, err := d.MsgCreator.New(mp.Head.MsgId)
	if err != nil {
		return gtp.MsgPacket{}, fmt.Errorf("gtp: new msg failed, %w (%d)", err, mp.Head.MsgId)
	}

	// 读取消息
	if _, err = msg.Write(msgBuf); err != nil {
		return gtp.MsgPacket{}, fmt.Errorf("gtp: read msg failed, %w", err)
	}

	mp.Msg = msg

	return mp, nil
}
