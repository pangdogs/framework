package codec

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/utils/binaryutil"
)

var encoder = MakeEncoder()

// DefaultEncoder 默认消息包编码器
func DefaultEncoder() Encoder {
	return encoder
}

// MakeEncoder 创建消息包编码器
func MakeEncoder() Encoder {
	return Encoder{}
}

// Encoder 消息包编码器
type Encoder struct{}

// Encode 编码消息包
func (Encoder) Encode(svc, src string, seq int64, msg gap.MsgReader) (ret binaryutil.RecycleBytes, err error) {
	if msg == nil {
		return binaryutil.NilRecycleBytes, fmt.Errorf("gap: %w: msg is nil", core.ErrArgs)
	}

	mp := gap.MsgPacket{
		Head: gap.MsgHead{
			MsgId: msg.MsgId(),
			Svc:   svc,
			Src:   src,
			Seq:   seq,
		},
		Msg: msg,
	}
	mp.Head.Len = uint32(mp.Size())

	mpBuf := binaryutil.MakeRecycleBytes(int(mp.Head.Len))
	defer func() {
		if !mpBuf.Equal(ret) {
			mpBuf.Release()
		}
	}()

	if _, err := mp.Read(mpBuf.Data()); err != nil {
		return binaryutil.NilRecycleBytes, err
	}

	return mpBuf, nil
}
