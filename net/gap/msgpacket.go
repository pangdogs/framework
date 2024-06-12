package gap

import "git.golaxy.org/framework/util/binaryutil"

// MsgPacket 消息包
type MsgPacket struct {
	Head MsgHead   // 消息头
	Msg  MsgReader // 消息
}

// Read implements io.Reader
func (mp MsgPacket) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	if _, err := binaryutil.ReadFrom(&bs, mp.Head); err != nil {
		return bs.BytesWritten(), err
	}

	if mp.Msg == nil {
		return bs.BytesWritten(), nil
	}

	if _, err := binaryutil.ReadFrom(&bs, mp.Msg); err != nil {
		return bs.BytesWritten(), err
	}

	return bs.BytesWritten(), nil
}

// Size 大小
func (mp MsgPacket) Size() int {
	n := mp.Head.Size()

	if mp.Msg != nil {
		n += mp.Msg.Size()
	}

	return n
}
