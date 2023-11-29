package codec

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"kit.golaxy.org/golaxy"
	"kit.golaxy.org/plugins/gtp"
	"kit.golaxy.org/plugins/util/binaryutil"
)

// IEncoder 消息包编码器接口
type IEncoder interface {
	io.Reader
	io.WriterTo
	// Reset 重置缓存
	Reset()
	// Encode 编码消息包
	Encode(flags gtp.Flags, msg gtp.MsgReader) error
	// EncodeTo 编码消息包，写入指定目标
	EncodeTo(writer io.Writer, flags gtp.Flags, msg gtp.MsgReader) error
	// GetEncryptionModule 获取加密模块
	GetEncryptionModule() IEncryptionModule
	// GetMACModule 获取MAC模块
	GetMACModule() IMACModule
	// GetCompressionModule 获取压缩模块
	GetCompressionModule() ICompressionModule
	// GetEncryption 获取开启加密
	GetEncryption() bool
	// GetPatchMAC 获取开启MAC
	GetPatchMAC() bool
	// GetCompressedSize 获取启用压缩阀值（字节），<=0表示不开启
	GetCompressedSize() int
}

// Encoder 消息包编码器
type Encoder struct {
	EncryptionModule  IEncryptionModule  // 加密模块
	MACModule         IMACModule         // MAC模块
	CompressionModule ICompressionModule // 压缩模块
	Encryption        bool               // 开启加密
	PatchMAC          bool               // 开启MAC
	CompressedSize    int                // 启用压缩阀值（字节），<=0表示不开启
	buffer            bytes.Buffer       // buffer
}

// Read implements io.Reader
func (e *Encoder) Read(p []byte) (int, error) {
	return e.buffer.Read(p)
}

// WriteTo implements io.WriterTo
func (e *Encoder) WriteTo(w io.Writer) (int64, error) {
	if w == nil {
		return 0, fmt.Errorf("%w: w is nil", golaxy.ErrArgs)
	}
	return e.buffer.WriteTo(w)
}

// Reset 重置缓存
func (e *Encoder) Reset() {
	e.buffer.Reset()
}

// Encode 编码消息包
func (e *Encoder) Encode(flags gtp.Flags, msg gtp.MsgReader) error {
	return e.EncodeTo(&e.buffer, flags, msg)
}

// EncodeTo 编码消息包，写入指定目标
func (e *Encoder) EncodeTo(writer io.Writer, flags gtp.Flags, msg gtp.MsgReader) error {
	if writer == nil {
		return fmt.Errorf("%w: writer is nil", golaxy.ErrArgs)
	}

	if msg == nil {
		return fmt.Errorf("%w: msg is nil", golaxy.ErrArgs)
	}

	head := gtp.MsgHead{}
	head.MsgId = msg.MsgId()

	head.Flags = flags.Setd(gtp.Flag_Encrypted, false).
		Setd(gtp.Flag_MAC, false).
		Setd(gtp.Flag_Compressed, false)

	// 预估追加的数据大小，因为后续数据可能会被压缩，所以此为评估值，只要保证不会内存溢出即可
	msgAddition := 0

	if e.Encryption {
		if e.EncryptionModule == nil {
			return errors.New("setting EncryptionModule is nil, msg can't be encrypted")
		}
		encAddition, err := e.EncryptionModule.SizeOfAddition(msg.Size())
		if err != nil {
			return fmt.Errorf("encrypt SizeOfAddition failed, %w", err)
		}
		msgAddition += encAddition

		if e.PatchMAC {
			if e.MACModule == nil {
				return errors.New("setting MACModule is nil, msg can't be patch MAC")
			}
			msgAddition += e.MACModule.SizeofMAC(msg.Size() + encAddition)
		}
	}

	mpBuf := binaryutil.MakeRecycleBytes(binaryutil.BytesPool.Get(head.Size() + msg.Size() + msgAddition))
	defer mpBuf.Release()

	// 写入消息
	mn, err := msg.Read(mpBuf.Data()[head.Size():])
	if err != nil {
		return fmt.Errorf("write msg failed, %w", err)
	}
	end := head.Size() + mn

	// 消息长度达到阀值，需要压缩消息
	if e.CompressedSize > 0 && msg.Size() >= e.CompressedSize {
		if e.CompressionModule == nil {
			return errors.New("setting CompressionModule is nil, msg can't be compress")
		}
		buf, compressed, err := e.CompressionModule.Compress(mpBuf.Data()[head.Size():end])
		if err != nil {
			return fmt.Errorf("compress msg failed, %w", err)
		}
		defer buf.Release()
		if compressed {
			head.Flags.Set(gtp.Flag_Compressed, true)

			copy(mpBuf.Data()[head.Size():], buf.Data())
			end = head.Size() + len(buf.Data())
		}
	}

	// 加密消息
	if e.Encryption {
		head.Flags.Set(gtp.Flag_Encrypted, true)

		// 补充MAC
		if e.PatchMAC {
			head.Flags.Set(gtp.Flag_MAC, true)

			if _, err = head.Read(mpBuf.Data()); err != nil {
				return fmt.Errorf("failed to write msg-packet-head for patch msg-mac, %w", err)
			}

			buf, err := e.MACModule.PatchMAC(head.MsgId, head.Flags, mpBuf.Data()[head.Size():end])
			if err != nil {
				return fmt.Errorf("patch msg-mac failed, %w", err)
			}
			defer buf.Release()

			copy(mpBuf.Data()[head.Size():], buf.Data())
			end = head.Size() + len(buf.Data())
		}

		buf, err := e.EncryptionModule.Transforming(mpBuf.Data()[head.Size():end], mpBuf.Data()[head.Size():end])
		if err != nil {
			return fmt.Errorf("encrypt msg failed, %w", err)
		}
		defer buf.Release()

		copy(mpBuf.Data()[head.Size():], buf.Data())
		end = head.Size() + len(buf.Data())
	}

	// 写入消息头
	head.Len = uint32(end)
	if _, err = head.Read(mpBuf.Data()); err != nil {
		return fmt.Errorf("write msg-packet-head failed, %w", err)
	}

	_, err = writer.Write(mpBuf.Data()[:end])
	if err != nil {
		return fmt.Errorf("write msg-packet-bytes failed, %w", err)
	}

	return nil
}

// GetEncryptionModule 获取加密模块
func (e *Encoder) GetEncryptionModule() IEncryptionModule {
	return e.EncryptionModule
}

// GetMACModule 获取MAC模块
func (e *Encoder) GetMACModule() IMACModule {
	return e.MACModule
}

// GetCompressionModule 获取压缩模块
func (e *Encoder) GetCompressionModule() ICompressionModule {
	return e.CompressionModule
}

// GetEncryption 获取开启加密
func (e *Encoder) GetEncryption() bool {
	return e.Encryption
}

// GetPatchMAC 获取开启MAC
func (e *Encoder) GetPatchMAC() bool {
	return e.PatchMAC
}

// GetCompressedSize 获取启用压缩阀值（字节），<=0表示不开启
func (e *Encoder) GetCompressedSize() int {
	return e.CompressedSize
}
