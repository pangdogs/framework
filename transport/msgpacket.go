package transport

import (
	"errors"
	"fmt"
	"io"
)

// MsgPacket 消息包
type MsgPacket struct {
	Head MsgHead
	Msg  Msg
}

func (p *MsgPacket) ReadFrom(r io.Reader) (int64, error) {
	hn, err := io.CopyN(&p.Head, r, int64(p.Head.Size()))
	if err != nil {
		return hn, err
	}

	msgLen := int64(p.Head.Len) - int64(p.Head.Size())
	var msg Msg

	switch p.Head.MsgId {
	case MsgId_Hello:
		msg = &MsgHello{}
	case MsgId_SecretKeyExchange:
		msg = &MsgECDHESecretKeyExchange{}
	case MsgId_ChangeCipherSpec:
		msg = &MsgChangeCipherSpec{}
	case MsgId_Auth:
		msg = &MsgAuth{}
	case MsgId_Finished:
		msg = &MsgFinished{}
	case MsgId_Rst:
		msg = &MsgRst{}
	case MsgId_Heartbeat:
		msg = &MsgHeartbeat{}
	case MsgId_SyncTime:
		msg = &MsgSyncTime{}
	case MsgId_Payload:
		msg = &MsgPayload{}
	default:
		mn, _ := io.CopyN(io.Discard, r, msgLen)
		return hn + mn, fmt.Errorf("msg %d not supported", p.Head.MsgId)
	}

	mn, err := io.CopyN(msg, r, msgLen)
	if err != nil {
		return hn + mn, err
	}
	p.Msg = msg

	return hn + mn, nil
}

func (p *MsgPacket) WriteTo(w io.Writer) (n int64, err error) {
	if p.Msg == nil {
		return 0, errors.New("msg is nil")
	}

	p.Head.Len = uint32(p.Head.Size()) + uint32(p.Msg.Size())
	p.Head.MsgId = p.Msg.MsgId()

	hn, err := io.CopyN(w, &p.Head, int64(p.Head.Size()))
	if err != nil {
		return hn, err
	}

	mn, err := io.CopyN(w, p.Msg, int64(p.Msg.Size()))
	if err != nil {
		return hn + mn, err
	}

	return hn + mn, nil
}
