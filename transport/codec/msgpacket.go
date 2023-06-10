package codec

import (
	"kit.golaxy.org/plugins/transport"
)

// MsgPacket 消息包
type MsgPacket struct {
	Head transport.MsgHead // 消息头
	Msg  transport.Msg     // 消息
}

func (mp *MsgPacket) Read(p []byte) (int, error) {
	hn, err := mp.Head.Read(p)
	if err != nil {
		return hn, err
	}
	if mp.Msg == nil {
		return hn, nil
	}
	mn, err := mp.Msg.Read(p)
	return mn + hn, err
}

func (mp *MsgPacket) Write(p []byte) (int, error) {
	hn, err := mp.Head.Write(p)
	if err != nil {
		return hn, err
	}
	if mp.Msg == nil {
		return hn, nil
	}
	mn, err := mp.Msg.Write(p)
	return mn + hn, err
}

func (mp *MsgPacket) Size() int {
	if mp.Msg == nil {
		return mp.Head.Size()
	}
	return mp.Head.Size() + mp.Msg.Size()
}
