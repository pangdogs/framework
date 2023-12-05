package codec

import (
	"bytes"
	"fmt"
	"io"
	"kit.golaxy.org/golaxy"
	"kit.golaxy.org/plugins/gap"
	"kit.golaxy.org/plugins/util/binaryutil"
)

// DefaultEncoder 默认消息包编码器
func DefaultEncoder() Encoder {
	return MakeEncoder()
}

// MakeEncoder 创建消息包编码器
func MakeEncoder() Encoder {
	return Encoder{}
}

// Encoder 消息包编码器
type Encoder struct {
	buffer bytes.Buffer // buffer
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

// Encode 编码消息包，写入缓存
func (e *Encoder) Encode(src string, seq int64, msg gap.Msg) error {
	return e.EncodeWriter(&e.buffer, src, seq, msg)
}

// EncodeWriter 编码消息包，写入指定writer
func (e Encoder) EncodeWriter(writer io.Writer, src string, seq int64, msg gap.Msg) error {
	if writer == nil {
		return fmt.Errorf("%w: writer is nil", golaxy.ErrArgs)
	}

	mpBuf, err := e.encode(src, seq, msg)
	if err != nil {
		return err
	}
	defer mpBuf.Release()

	_, err = writer.Write(mpBuf.Data())
	if err != nil {
		return fmt.Errorf("write msg-packet-bytes failed, %w", err)
	}

	return nil
}

// EncodeBuff 编码消息包，写入指定buffer
func (e Encoder) EncodeBuff(buff *bytes.Buffer, src string, seq int64, msg gap.Msg) error {
	if buff == nil {
		return fmt.Errorf("%w: buff is nil", golaxy.ErrArgs)
	}
	return e.EncodeWriter(buff, src, seq, msg)
}

// EncodeBytes 编码消息包，返回可回收bytes
func (e Encoder) EncodeBytes(src string, seq int64, msg gap.Msg) (binaryutil.RecycleBytes, error) {
	return e.encode(src, seq, msg)
}

// encode 编码消息包
func (Encoder) encode(src string, seq int64, msg gap.Msg) (ret binaryutil.RecycleBytes, err error) {
	if msg == nil {
		return binaryutil.MakeNonRecycleBytes(nil), fmt.Errorf("%w: msg is nil", golaxy.ErrArgs)
	}

	mp := gap.MsgPacket{
		Head: gap.MsgHead{
			MsgId: msg.MsgId(),
			Src:   src,
			Seq:   seq,
		},
		Msg: msg,
	}
	mp.Head.Len = uint32(mp.Size())

	mpBuf := binaryutil.MakeRecycleBytes(binaryutil.BytesPool.Get(int(mp.Head.Len)))
	defer func() {
		if err != nil {
			mpBuf.Release()
		}
	}()

	if _, err := mp.Read(mpBuf.Data()); err != nil {
		return binaryutil.MakeNonRecycleBytes(nil), err
	}

	return mpBuf, nil
}
