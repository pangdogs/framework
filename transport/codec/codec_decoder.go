package codec

import (
	"bytes"
	"errors"
	"io"
	"kit.golaxy.org/plugins/transport"
)

var (
	ErrEmptyBuffer      = errors.New("i/o empty buffer")   // 缓存空
	ErrMsgNotRegistered = errors.New("msg not registered") // 消息未注册
)

// IDecoder 消息包解码器接口
type IDecoder interface {
	io.Writer
	io.ReaderFrom
	// Fetch 取出单个消息包
	Fetch(fun func(mp transport.MsgPacket)) error
	// MultiFetch 取出多个消息包
	MultiFetch(fun func(mp transport.MsgPacket) bool) error
	// Discard 丢弃消息包
	Discard(n int) error
	// GetMsgCreator 获取消息构建器
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
	MsgCreator        IMsgCreator        // 消息构建器
	EncryptionModule  IEncryptionModule  // 加密模块
	MACModule         IMACModule         // MAC模块
	CompressionModule ICompressionModule // 压缩模块
	buffer            bytes.Buffer       // buffer
	gcList            [][]byte           // GC列表
}

func (d *Decoder) Write(p []byte) (int, error) {
	return d.buffer.Write(p)
}

func (d *Decoder) ReadFrom(r io.Reader) (int64, error) {
	var buff [bytes.MinRead]byte

	n, err := r.Read(buff[:])
	if n > 0 {
		d.buffer.Write(buff[:n])
	}

	return int64(n), err
}

// Fetch 取出单个消息包
func (d *Decoder) Fetch(fun func(mp transport.MsgPacket)) error {
	if fun == nil {
		return errors.New("fun is nil")
	}

	mpl := transport.MsgPacketLen{}

	// 读取消息包长度
	_, err := mpl.Write(d.buffer.Bytes())
	if err != nil {
		return ErrEmptyBuffer
	}

	if d.buffer.Len() < int(mpl.Len) {
		return ErrEmptyBuffer
	}

	buf := BytesPool.Get(int(mpl.Len))
	d.gcList = append(d.gcList, buf)

	// 读取消息包
	_, err = d.buffer.Read(buf)
	if err != nil {
		return err
	}

	mp := transport.MsgPacket{}

	// 读取消息头
	_, err = mp.Head.Write(buf)
	if err != nil {
		return err
	}

	// 创建消息体
	mp.Msg, err = d.MsgCreator.Spawn(mp.Head.MsgId)
	if err != nil {
		return errors.Join(ErrMsgNotRegistered, err)
	}

	msgBuf := buf[mp.Head.Size():]

	// 检查加密标记
	if mp.Head.Flags.Is(transport.Flag_Encrypted) {
		// 解密消息体
		if d.EncryptionModule == nil {
			return errors.New("setting EncryptionModule is nil, msg can't be decrypted")
		}
		msgBuf, err = d.EncryptionModule.Transforming(msgBuf, msgBuf)
		if err != nil {
			return err
		}

		// 检查MAC标记
		if mp.Head.Flags.Is(transport.Flag_MAC) {
			if d.MACModule == nil {
				return errors.New("setting MACModule is nil, msg can't be verify MAC")
			}
			// 检测MAC
			msgBuf, err = d.MACModule.VerifyMAC(buf[:mp.Head.Size()], msgBuf)
			if err != nil {
				return err
			}
		}
	}

	// 检查压缩标记
	if mp.Head.Flags.Is(transport.Flag_Compressed) {
		if d.CompressionModule == nil {
			return errors.New("setting CompressionModule is nil, msg can't be uncompress")
		}
		msgBuf, err = d.CompressionModule.Uncompress(msgBuf)
		if err != nil {
			return err
		}
	}

	// 读取消息
	_, err = mp.Msg.Write(msgBuf)
	if err != nil {
		return err
	}

	fun(mp)

	return nil
}

// MultiFetch 取出多个消息包
func (d *Decoder) MultiFetch(fun func(mp transport.MsgPacket) bool) error {
	if fun == nil {
		return errors.New("fun is nil")
	}

	for {
		var recvMP transport.MsgPacket

		if err := d.Fetch(func(mp transport.MsgPacket) { recvMP = mp }); err != nil {
			return err
		}

		if !fun(recvMP) {
			return nil
		}
	}
}

// Discard 丢弃消息包
func (d *Decoder) Discard(n int) error {
	for i := 0; i < n; i++ {
		mpl := transport.MsgPacketLen{}

		// 读取消息包长度
		_, err := mpl.Write(d.buffer.Bytes())
		if err != nil {
			return ErrEmptyBuffer
		}

		if d.buffer.Len() < int(mpl.Len) {
			return ErrEmptyBuffer
		}

		d.buffer.Next(int(mpl.Len))
	}
	return nil
}

// GetMsgCreator 获取消息构建器
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
