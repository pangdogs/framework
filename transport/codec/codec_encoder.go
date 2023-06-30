package codec

import (
	"bytes"
	"errors"
	"io"
	"kit.golaxy.org/plugins/transport"
)

// IEncoder 消息包编码器接口
type IEncoder interface {
	io.Reader
	io.WriterTo
	// Stuff 填充消息
	Stuff(flags transport.Flags, msg transport.Msg) error
}

// Encoder 消息包编码器
type Encoder struct {
	EncryptionModule  IEncryptionModule  // 加密模块
	MACModule         IMACModule         // MAC模块
	CompressionModule ICompressionModule // 压缩模块
	Encryption        bool               // 开启加密
	PatchMAC          bool               // 开启MAC
	Compressed        int                // 开启压缩阀值，<=0表示不开启
	buffer            bytes.Buffer       // buffer
}

func (e *Encoder) Read(p []byte) (int, error) {
	return e.buffer.Read(p)
}

func (e *Encoder) WriteTo(w io.Writer) (int64, error) {
	return e.buffer.WriteTo(w)
}

// Stuff 填充消息
func (e *Encoder) Stuff(flags transport.Flags, msg transport.Msg) error {
	if msg == nil {
		return errors.New("msg is nil")
	}

	head := transport.MsgHead{}
	head.Flags = (flags << transport.Flag_Customize) >> transport.Flag_Customize
	head.MsgId = msg.MsgId()

	// 预估追加的数据大小，因为后续数据可能会被压缩，所以此为评估值，只要保证不会内存溢出即可
	msgAddition := 0

	if e.Encryption {
		if e.EncryptionModule == nil {
			return errors.New("setting EncryptionModule is nil, msg can't be encrypted")
		}
		encAddition, err := e.EncryptionModule.SizeOfAddition(msg.Size())
		if err != nil {
			return err
		}
		msgAddition += encAddition

		if e.PatchMAC {
			if e.MACModule == nil {
				return errors.New("setting MACModule is nil, msg can't be patch MAC")
			}
			msgAddition += e.MACModule.SizeofMAC(msg.Size() + encAddition)
		}
	}

	mpBuf := BytesPool.Get(head.Size() + msg.Size() + msgAddition)
	defer BytesPool.Put(mpBuf)

	// 写入消息
	mn, err := msg.Read(mpBuf[head.Size():])
	if err != nil {
		return err
	}
	end := head.Size() + mn

	// 消息长度达到阀值，需要压缩消息
	if e.Compressed > 0 && msg.Size() >= e.Compressed {
		if e.CompressionModule == nil {
			return errors.New("setting CompressionModule is nil, msg can't be compress")
		}
		buf, compressed, err := e.CompressionModule.Compress(mpBuf[head.Size():end])
		if err != nil {
			return err
		}
		if compressed {
			defer e.CompressionModule.GC()

			head.Flags.Set(transport.Flag_Compressed, true)

			copy(mpBuf[head.Size():], buf)
			end = head.Size() + len(buf)
		}
	}

	// 加密消息
	if e.Encryption {
		head.Flags.Set(transport.Flag_Encrypted, true)

		// 补充MAC
		if e.PatchMAC {
			head.Flags.Set(transport.Flag_MAC, true)

			if _, err = head.Read(mpBuf); err != nil {
				return err
			}

			buf, err := e.MACModule.PatchMAC(mpBuf[:head.Size()], mpBuf[head.Size():end])
			if err != nil {
				return err
			}
			defer e.MACModule.GC()

			copy(mpBuf[head.Size():], buf)
			end = head.Size() + len(buf)
		}

		buf, err := e.EncryptionModule.Transforming(mpBuf[head.Size():end], mpBuf[head.Size():end])
		if err != nil {
			return err
		}
		defer e.EncryptionModule.GC()

		copy(mpBuf[head.Size():], buf)
		end = head.Size() + len(buf)
	}

	// 写入消息头
	head.Len = uint32(end)
	if _, err = head.Read(mpBuf); err != nil {
		return err
	}

	// 写入缓冲
	_, err = e.buffer.Write(mpBuf[:end])
	return err
}
