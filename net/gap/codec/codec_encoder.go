package codec

import (
	"bytes"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/util/binaryutil"
	"io"
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
		return 0, fmt.Errorf("gap: %w: w is nil", core.ErrArgs)
	}
	return e.buffer.WriteTo(w)
}

// Reset 重置缓存
func (e *Encoder) Reset() {
	e.buffer.Reset()
}

// Encode 编码消息包，写入缓存
func (e *Encoder) Encode(src string, seq int64, msg gap.MsgReader) error {
	return e.EncodeWriter(&e.buffer, src, seq, msg)
}

// EncodeWriter 编码消息包，写入指定writer
func (e Encoder) EncodeWriter(writer io.Writer, src string, seq int64, msg gap.MsgReader) error {
	if writer == nil {
		return fmt.Errorf("gap: %w: writer is nil", core.ErrArgs)
	}

	mpBuf, err := e.encode(src, seq, msg)
	if err != nil {
		return err
	}
	defer mpBuf.Release()

	_, err = writer.Write(mpBuf.Data())
	if err != nil {
		return fmt.Errorf("gap: write msg-packet failed, %w", err)
	}

	return nil
}

// EncodeBuff 编码消息包，写入指定buffer
func (e Encoder) EncodeBuff(buff *bytes.Buffer, src string, seq int64, msg gap.MsgReader) error {
	if buff == nil {
		return fmt.Errorf("gap: %w: buff is nil", core.ErrArgs)
	}
	return e.EncodeWriter(buff, src, seq, msg)
}

// EncodeBytes 编码消息包，返回可回收bytes
func (e Encoder) EncodeBytes(src string, seq int64, msg gap.MsgReader) (binaryutil.RecycleBytes, error) {
	return e.encode(src, seq, msg)
}

// encode 编码消息包
func (Encoder) encode(src string, seq int64, msg gap.MsgReader) (ret binaryutil.RecycleBytes, err error) {
	if msg == nil {
		return binaryutil.NilRecycleBytes, fmt.Errorf("gap: %w: msg is nil", core.ErrArgs)
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
		return binaryutil.NilRecycleBytes, err
	}

	return mpBuf, nil
}
