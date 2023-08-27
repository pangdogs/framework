package codec

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"kit.golaxy.org/plugins/gtp"
)

var (
	ErrBufferNotEnough  = errors.New("buffer data not enough") // 缓冲区数据不足
	ErrMsgNotRegistered = errors.New("msg not registered")     // 消息未注册
)

type (
	Validation = func(msgHead gtp.MsgHead) error // 验证消息头
)

// IDecoder 消息包解码器接口
type IDecoder interface {
	io.Writer
	io.ReaderFrom
	// Reset 重置缓存
	Reset()
	// Fetch 取出消息包
	Fetch(validation Validation) (gtp.MsgPacket, error)
	// FetchFrom 取出消息包
	FetchFrom(buff *bytes.Buffer, validation Validation) (gtp.MsgPacket, error)
	// GetMsgCreator 获取消息对象构建器
	GetMsgCreator() IMsgCreator
	// GetEncryptionModule 获取加密模块
	GetEncryptionModule() IEncryptionModule
	// GetMACModule 获取MAC模块
	GetMACModule() IMACModule
	// GetCompressionModule 获取压缩模块
	GetCompressionModule() ICompressionModule
	// GC GC
	GC()
}

// Decoder 消息包解码器
type Decoder struct {
	MsgCreator        IMsgCreator        // 消息对象构建器
	EncryptionModule  IEncryptionModule  // 加密模块
	MACModule         IMACModule         // MAC模块
	CompressionModule ICompressionModule // 压缩模块
	buffer            bytes.Buffer       // buffer
	gcList            [][]byte           // GC列表
}

// Write implements io.Writer
func (d *Decoder) Write(p []byte) (int, error) {
	return d.buffer.Write(p)
}

// ReadFrom implements io.ReaderFrom
func (d *Decoder) ReadFrom(r io.Reader) (int64, error) {
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

// Fetch 取出消息包
func (d *Decoder) Fetch(validation Validation) (gtp.MsgPacket, error) {
	return d.FetchFrom(&d.buffer, validation)
}

// FetchFrom 取出消息包
func (d *Decoder) FetchFrom(buff *bytes.Buffer, validation Validation) (gtp.MsgPacket, error) {
	mpl := gtp.MsgPacketLen{}

	// 读取消息包长度
	_, err := mpl.Write(buff.Bytes())
	if err != nil {
		return gtp.MsgPacket{}, ErrBufferNotEnough
	}

	if buff.Len() < int(mpl.Len) {
		return gtp.MsgPacket{}, fmt.Errorf("%w: %d < %d", ErrBufferNotEnough, buff.Len(), mpl.Len)
	}

	buf := BytesPool.Get(int(mpl.Len))
	d.gcList = append(d.gcList, buf)

	// 读取消息包
	_, err = buff.Read(buf)
	if err != nil {
		return gtp.MsgPacket{}, fmt.Errorf("read msg-packet-bytes failed, %w", err)
	}

	mp := gtp.MsgPacket{}

	// 读取消息头
	_, err = mp.Head.Write(buf)
	if err != nil {
		return gtp.MsgPacket{}, fmt.Errorf("read msg-packet-head failed, %w", err)
	}

	// 验证消息头
	if validation != nil {
		err = validation(mp.Head)
		if err != nil {
			return gtp.MsgPacket{}, fmt.Errorf("validation msg-packet-head failed, %w", err)
		}
	}

	// 创建消息体
	mp.Msg, err = d.MsgCreator.Spawn(mp.Head.MsgId)
	if err != nil {
		return gtp.MsgPacket{}, fmt.Errorf("%w, %w", ErrMsgNotRegistered, err)
	}

	msgBuf := buf[mp.Head.Size():]

	// 检查加密标记
	if mp.Head.Flags.Is(gtp.Flag_Encrypted) {
		// 解密消息体
		if d.EncryptionModule == nil {
			return gtp.MsgPacket{}, errors.New("setting EncryptionModule is nil, msg can't be decrypted")
		}
		msgBuf, err = d.EncryptionModule.Transforming(msgBuf, msgBuf)
		if err != nil {
			return gtp.MsgPacket{}, fmt.Errorf("dencrypt msg failed, %w", err)
		}

		// 检查MAC标记
		if mp.Head.Flags.Is(gtp.Flag_MAC) {
			if d.MACModule == nil {
				return gtp.MsgPacket{}, errors.New("setting MACModule is nil, msg can't be verify MAC")
			}
			// 检测MAC
			msgBuf, err = d.MACModule.VerifyMAC(mp.Head.MsgId, mp.Head.Flags, msgBuf)
			if err != nil {
				return gtp.MsgPacket{}, fmt.Errorf("verify msg-mac failed, %w", err)
			}
		}
	}

	// 检查压缩标记
	if mp.Head.Flags.Is(gtp.Flag_Compressed) {
		if d.CompressionModule == nil {
			return gtp.MsgPacket{}, errors.New("setting CompressionModule is nil, msg can't be uncompress")
		}
		msgBuf, err = d.CompressionModule.Uncompress(msgBuf)
		if err != nil {
			return gtp.MsgPacket{}, fmt.Errorf("uncompress msg failed, %w", err)
		}
	}

	// 读取消息
	_, err = mp.Msg.Write(msgBuf)
	if err != nil {
		return gtp.MsgPacket{}, fmt.Errorf("read msg failed, %w", err)
	}

	return mp, nil
}

// GetMsgCreator 获取消息对象构建器
func (d *Decoder) GetMsgCreator() IMsgCreator {
	return d.MsgCreator
}

// GetEncryptionModule 获取加密模块
func (d *Decoder) GetEncryptionModule() IEncryptionModule {
	return d.EncryptionModule
}

// GetMACModule 获取MAC模块
func (d *Decoder) GetMACModule() IMACModule {
	return d.MACModule
}

// GetCompressionModule 获取压缩模块
func (d *Decoder) GetCompressionModule() ICompressionModule {
	return d.CompressionModule
}

// GC GC
func (d *Decoder) GC() {
	for i := range d.gcList {
		BytesPool.Put(d.gcList[i])
	}
	d.gcList = d.gcList[:0]

	if d.EncryptionModule != nil {
		d.EncryptionModule.GC()
	}

	if d.MACModule != nil {
		d.MACModule.GC()
	}

	if d.CompressionModule != nil {
		d.CompressionModule.GC()
	}
}
