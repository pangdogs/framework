package codec

import (
	"bytes"
	"errors"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/plugins/gtp"
	"git.golaxy.org/plugins/util/binaryutil"
	"io"
)

// IEncoder 消息包编码器接口
type IEncoder interface {
	io.Reader
	io.WriterTo
	// Reset 重置缓存
	Reset()
	// Encode 编码消息包，写入缓存
	Encode(flags gtp.Flags, msg gtp.MsgReader) error
	// EncodeWriter 编码消息包，写入指定writer
	EncodeWriter(writer io.Writer, flags gtp.Flags, msg gtp.MsgReader) error
	// EncodeBuff 编码消息包，写入指定buffer
	EncodeBuff(buff *bytes.Buffer, flags gtp.Flags, msg gtp.MsgReader) error
	// EncodeBytes 编码消息包，返回可回收bytes
	EncodeBytes(flags gtp.Flags, msg gtp.MsgReader) (binaryutil.RecycleBytes, error)
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
		return 0, fmt.Errorf("gtp: %w: w is nil", core.ErrArgs)
	}
	return e.buffer.WriteTo(w)
}

// Reset 重置缓存
func (e *Encoder) Reset() {
	e.buffer.Reset()
}

// Encode 编码消息包，写入缓存
func (e *Encoder) Encode(flags gtp.Flags, msg gtp.MsgReader) error {
	return e.EncodeWriter(&e.buffer, flags, msg)
}

// EncodeWriter 编码消息包，写入指定writer
func (e *Encoder) EncodeWriter(writer io.Writer, flags gtp.Flags, msg gtp.MsgReader) error {
	if writer == nil {
		return fmt.Errorf("gtp: %w: writer is nil", core.ErrArgs)
	}

	mpBuf, err := e.encode(flags, msg)
	if err != nil {
		return err
	}
	defer mpBuf.Release()

	_, err = writer.Write(mpBuf.Data())
	if err != nil {
		return fmt.Errorf("gtp: write msg-packet failed, %w", err)
	}

	return nil
}

// EncodeBuff 编码消息包，写入指定buffer
func (e *Encoder) EncodeBuff(buff *bytes.Buffer, flags gtp.Flags, msg gtp.MsgReader) error {
	if buff == nil {
		return fmt.Errorf("gtp: %w: buff is nil", core.ErrArgs)
	}
	return e.EncodeWriter(buff, flags, msg)
}

// EncodeBytes 编码消息包，返回可回收bytes
func (e *Encoder) EncodeBytes(flags gtp.Flags, msg gtp.MsgReader) (binaryutil.RecycleBytes, error) {
	return e.encode(flags, msg)
}

// encode 编码消息包
func (e *Encoder) encode(flags gtp.Flags, msg gtp.MsgReader) (ret binaryutil.RecycleBytes, err error) {
	if msg == nil {
		return binaryutil.MakeNonRecycleBytes(nil), fmt.Errorf("gtp: %w: msg is nil", core.ErrArgs)
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
			return binaryutil.MakeNonRecycleBytes(nil), errors.New("gtp: setting EncryptionModule is nil, msg can't be encrypted")
		}
		encAddition, err := e.EncryptionModule.SizeOfAddition(msg.Size())
		if err != nil {
			return binaryutil.MakeNonRecycleBytes(nil), fmt.Errorf("gtp: encrypt SizeOfAddition failed, %w", err)
		}
		msgAddition += encAddition

		if e.PatchMAC {
			if e.MACModule == nil {
				return binaryutil.MakeNonRecycleBytes(nil), errors.New("gtp: setting MACModule is nil, msg can't be patch MAC")
			}
			msgAddition += e.MACModule.SizeofMAC(msg.Size() + encAddition)
		}
	}

	mpBuf := binaryutil.MakeRecycleBytes(binaryutil.BytesPool.Get(head.Size() + msg.Size() + msgAddition))
	defer func() {
		if err != nil {
			mpBuf.Release()
		}
	}()

	// 写入消息
	mn, err := msg.Read(mpBuf.Data()[head.Size():])
	if err != nil {
		return binaryutil.MakeNonRecycleBytes(nil), fmt.Errorf("gtp: write msg failed, %w", err)
	}
	end := head.Size() + mn

	// 消息长度达到阀值，需要压缩消息
	if e.CompressedSize > 0 && msg.Size() >= e.CompressedSize {
		if e.CompressionModule == nil {
			return binaryutil.MakeNonRecycleBytes(nil), errors.New("gtp: setting CompressionModule is nil, msg can't be compress")
		}
		compressedBuf, compressed, err := e.CompressionModule.Compress(mpBuf.Data()[head.Size():end])
		if err != nil {
			return binaryutil.MakeNonRecycleBytes(nil), fmt.Errorf("gtp: compress msg failed, %w", err)
		}
		defer compressedBuf.Release()
		if compressed {
			head.Flags.Set(gtp.Flag_Compressed, true)

			copy(mpBuf.Data()[head.Size():], compressedBuf.Data())
			end = head.Size() + len(compressedBuf.Data())
		}
	}

	// 加密消息
	if e.Encryption {
		head.Flags.Set(gtp.Flag_Encrypted, true)

		// 补充MAC
		if e.PatchMAC {
			head.Flags.Set(gtp.Flag_MAC, true)

			if _, err = head.Read(mpBuf.Data()); err != nil {
				return binaryutil.MakeNonRecycleBytes(nil), fmt.Errorf("gtp: failed to write msg-packet-head for patch msg-mac, %w", err)
			}

			macBuf, err := e.MACModule.PatchMAC(head.MsgId, head.Flags, mpBuf.Data()[head.Size():end])
			if err != nil {
				return binaryutil.MakeNonRecycleBytes(nil), fmt.Errorf("gtp: patch msg-mac failed, %w", err)
			}
			defer macBuf.Release()

			copy(mpBuf.Data()[head.Size():], macBuf.Data())
			end = head.Size() + len(macBuf.Data())
		}

		// 加密消息体
		encryptBuf, err := e.EncryptionModule.Transforming(mpBuf.Data()[head.Size():end], mpBuf.Data()[head.Size():end])
		if err != nil {
			return binaryutil.MakeNonRecycleBytes(nil), fmt.Errorf("gtp: encrypt msg failed, %w", err)
		}
		defer encryptBuf.Release()

		copy(mpBuf.Data()[head.Size():], encryptBuf.Data())
		end = head.Size() + len(encryptBuf.Data())
	}

	// 调整消息大小
	mpBuf = binaryutil.SliceRecycleBytes(mpBuf, 0, end)

	// 写入消息头
	head.Len = uint32(len(mpBuf.Data()))
	if _, err = head.Read(mpBuf.Data()); err != nil {
		return binaryutil.MakeNonRecycleBytes(nil), fmt.Errorf("gtp: write msg-packet-head failed, %w", err)
	}

	return mpBuf, nil
}
