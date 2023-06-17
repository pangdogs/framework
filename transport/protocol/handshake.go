package tcp

import (
	"errors"
	"fmt"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/codec"
	"net"
)

type (
	HelloFin    = func(MsgBlock[*transport.MsgHello]) error
	HelloAccept = func(MsgBlock[*transport.MsgHello]) (MsgBlock[*transport.MsgHello], error)
)

// Handshake 服务端握手协议
type Handshake struct {
	Conn       net.Conn       // 网络连接
	Encoder    codec.IEncoder // 消息包编码器
	Decoder    codec.IDecoder // 消息包解码器
	RetryTimes int            // io超时重试次数
}

// ClientHello 客户端Hello
func (h Handshake) ClientHello(sendMB MsgBlock[*transport.MsgHello], fin HelloFin) error {
	if fin == nil {
		return errors.New("fin is nil")
	}

	trans := Transceiver{
		Conn:       h.Conn,
		Encoder:    h.Encoder,
		Decoder:    h.Decoder,
		RetryTimes: h.RetryTimes,
	}

	err := trans.Send(MsgBlock[transport.Msg]{
		Flags: sendMB.Flags,
		Msg:   sendMB.Msg,
	})
	if err != nil {
		return err
	}

	recvMB, err := trans.Recv()
	if err != nil {
		return err
	}

	switch recvMB.Msg.MsgId() {
	case transport.MsgId_Hello:
		err := fin(MsgBlock[*transport.MsgHello]{
			Flags: recvMB.Flags,
			Msg:   recvMB.Msg.(*transport.MsgHello),
		})
		if err != nil {
			return err
		}
	case transport.MsgId_Rst:
		msg := recvMB.Msg.(*transport.MsgRst)
		return fmt.Errorf("recv rst, %s", msg.Message)
	default:
		return fmt.Errorf("recv unexpected msg %d", recvMB.Msg.MsgId())
	}

	return nil
}

// ServerHello 服务端Hello
func (h Handshake) ServerHello(accept HelloAccept) (err error) {
	if accept == nil {
		return errors.New("accept is nil")
	}

	trans := Transceiver{
		Conn:       h.Conn,
		Encoder:    h.Encoder,
		Decoder:    h.Decoder,
		RetryTimes: h.RetryTimes,
	}

	defer func() {
		if err != nil {
			trans.SendRst(err)
		}
	}()

	recvMsg, err := trans.Recv()
	if err != nil {
		return err
	}

	if recvMsg.Msg.MsgId() != transport.MsgId_Hello {
		return fmt.Errorf("recv unexpected msg %d", recvMsg.Msg.MsgId())
	}

	sendMsg, err := accept(MsgBlock[*transport.MsgHello]{
		Flags: recvMsg.Flags,
		Msg:   recvMsg.Msg.(*transport.MsgHello),
	})
	if err != nil {
		return err
	}

	if err = trans.Send(MsgBlock[transport.Msg]{
		Flags: sendMsg.Flags,
		Msg:   sendMsg.Msg,
	}); err != nil {
		return err
	}

	return nil
}

func (h Handshake) ClientSecretKeyExchange() error {
	return nil
}

func (h Handshake) ServerSecretKeyExchange() error {
	return nil
}

func (h Handshake) ClientAuth() error {
	return nil
}

func (h Handshake) ServerAuth() error {
	return nil
}
