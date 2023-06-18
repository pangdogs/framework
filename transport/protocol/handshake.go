package tcp

import (
	"errors"
	"fmt"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/codec"
	"net"
)

type (
	HelloAccept               = func(Msg[*transport.MsgHello]) (Msg[*transport.MsgHello], error)
	HelloFin                  = func(Msg[*transport.MsgHello]) error
	SecretKeyExchangeAccept   = func(Msg[transport.Msg]) (Msg[transport.Msg], error)
	ECDHESecretKeyExchangeFin = func(Msg[*transport.MsgECDHESecretKeyExchange]) (Msg[*transport.MsgChangeCipherSpec], error)
	ChangeCipherSpecAccept    = func(Msg[*transport.MsgChangeCipherSpec]) (Msg[*transport.MsgChangeCipherSpec], error)
	ChangeCipherSpecFin       = func(Msg[*transport.MsgChangeCipherSpec]) error
	AuthAccept                = func(Msg[*transport.MsgAuth]) error
	FinishedAccept            = func(Msg[*transport.MsgFinished]) error
)

// Handshake 握手协议
type Handshake struct {
	Conn       net.Conn       // 网络连接
	Encoder    codec.IEncoder // 消息包编码器
	Decoder    codec.IDecoder // 消息包解码器
	RetryTimes int            // io超时重试次数
}

// ClientHello 客户端Hello
func (h *Handshake) ClientHello(helloMsg Msg[*transport.MsgHello], helloFin HelloFin) error {
	if helloFin == nil {
		return errors.New("helloFin is nil")
	}

	trans := Transceiver{
		Conn:       h.Conn,
		Encoder:    h.Encoder,
		Decoder:    h.Decoder,
		RetryTimes: h.RetryTimes,
	}

	defer trans.Decoder.GC()

	err := trans.Send(Msg2Trans(helloMsg))
	if err != nil {
		return err
	}

	recv, err := trans.Recv()
	if err != nil {
		return err
	}

	switch recv.Msg.MsgId() {
	case transport.MsgId_Hello:
		break
	case transport.MsgId_Rst:
		return fmt.Errorf("recv rst, %s", Trans2Msg[*transport.MsgRst](recv).Msg.Message)
	default:
		return fmt.Errorf("recv unexpected msg %d", recv.Msg.MsgId())
	}

	err = helloFin(Trans2Msg[*transport.MsgHello](recv))
	if err != nil {
		return err
	}

	return nil
}

// ServerHello 服务端Hello
func (h *Handshake) ServerHello(helloAccept HelloAccept) (err error) {
	if helloAccept == nil {
		return errors.New("helloAccept is nil")
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
		trans.Decoder.GC()
	}()

	recv, err := trans.Recv()
	if err != nil {
		return err
	}

	switch recv.Msg.MsgId() {
	case transport.MsgId_Hello:
		break
	default:
		return fmt.Errorf("recv unexpected msg %d", recv.Msg.MsgId())
	}

	reply, err := helloAccept(Trans2Msg[*transport.MsgHello](recv))
	if err != nil {
		return err
	}

	err = trans.Send(Msg2Trans(reply))
	if err != nil {
		return err
	}

	return nil
}

// ClientSecretKeyExchange 客户端交换秘钥
func (h *Handshake) ClientSecretKeyExchange(secretKeyExchangeAccept SecretKeyExchangeAccept, changeCipherSpecAccept ChangeCipherSpecAccept) error {
	if secretKeyExchangeAccept == nil {
		return errors.New("secretKeyExchangeAccept is nil")
	}

	if changeCipherSpecAccept == nil {
		return errors.New("changeCipherSpecAccept is nil")
	}

	trans := Transceiver{
		Conn:       h.Conn,
		Encoder:    h.Encoder,
		Decoder:    h.Decoder,
		RetryTimes: h.RetryTimes,
	}

	defer trans.Decoder.GC()

	recv, err := trans.Recv()
	if err != nil {
		return err
	}

	switch recv.Msg.MsgId() {
	case transport.MsgId_ECDHESecretKeyExchange:
		break
	case transport.MsgId_Rst:
		return fmt.Errorf("recv rst, %s", Trans2Msg[*transport.MsgRst](recv).Msg.Message)
	default:
		return fmt.Errorf("recv unexpected msg %d", recv.Msg.MsgId())
	}

	secretKeyExchangeReply, err := secretKeyExchangeAccept(recv)
	if err != nil {
		return err
	}

	err = trans.Send(Msg2Trans(secretKeyExchangeReply))
	if err != nil {
		return err
	}

	recv, err = trans.Recv()
	if err != nil {
		return err
	}

	switch recv.Msg.MsgId() {
	case transport.MsgId_ChangeCipherSpec:
		break
	case transport.MsgId_Rst:
		return fmt.Errorf("recv rst, %s", Trans2Msg[*transport.MsgRst](recv).Msg.Message)
	default:
		return fmt.Errorf("recv unexpected msg %d", recv.Msg.MsgId())
	}

	changeCipherSpecReply, err := changeCipherSpecAccept(Trans2Msg[*transport.MsgChangeCipherSpec](recv))
	if err != nil {
		return err
	}

	err = trans.Send(Msg2Trans(changeCipherSpecReply))
	if err != nil {
		return err
	}

	return nil
}

// ServerECDHESecretKeyExchange 服务端交换秘钥（ECDHE）
func (h *Handshake) ServerECDHESecretKeyExchange(secretKeyExchangeMsg Msg[*transport.MsgECDHESecretKeyExchange], secretKeyExchangeFin ECDHESecretKeyExchangeFin, changeCipherSpecFin ChangeCipherSpecFin) (err error) {
	if secretKeyExchangeFin == nil {
		return errors.New("secretKeyExchangeFin is nil")
	}

	if changeCipherSpecFin == nil {
		return errors.New("changeCipherSpecFin is nil")
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
		trans.Decoder.GC()
	}()

	err = trans.Send(Msg2Trans(secretKeyExchangeMsg))
	if err != nil {
		return err
	}

	recv, err := trans.Recv()
	if err != nil {
		return err
	}

	switch recv.Msg.MsgId() {
	case transport.MsgId_ECDHESecretKeyExchange:
		break
	default:
		return fmt.Errorf("recv unexpected msg %d", recv.Msg.MsgId())
	}

	changeCipherSpecMsg, err := secretKeyExchangeFin(Trans2Msg[*transport.MsgECDHESecretKeyExchange](recv))
	if err != nil {
		return err
	}

	err = trans.Send(Msg2Trans(changeCipherSpecMsg))
	if err != nil {
		return err
	}

	recv, err = trans.Recv()
	if err != nil {
		return err
	}

	switch recv.Msg.MsgId() {
	case transport.MsgId_ChangeCipherSpec:
		break
	default:
		return fmt.Errorf("recv unexpected msg %d", recv.Msg.MsgId())
	}

	err = changeCipherSpecFin(Trans2Msg[*transport.MsgChangeCipherSpec](recv))
	if err != nil {
		return err
	}

	return nil
}

// ClientAuth 客户端发起鉴权
func (h *Handshake) ClientAuth(authMsg Msg[*transport.MsgAuth]) error {
	trans := Transceiver{
		Conn:       h.Conn,
		Encoder:    h.Encoder,
		Decoder:    h.Decoder,
		RetryTimes: h.RetryTimes,
	}

	err := trans.Send(Msg2Trans(authMsg))
	if err != nil {
		return err
	}

	return nil
}

// ServerAuth 服务端验证鉴权
func (h *Handshake) ServerAuth(authAccept AuthAccept) (err error) {
	if authAccept == nil {
		return errors.New("authAccept is nil")
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
		trans.Decoder.GC()
	}()

	recv, err := trans.Recv()
	if err != nil {
		return err
	}

	switch recv.Msg.MsgId() {
	case transport.MsgId_Auth:
		break
	default:
		return fmt.Errorf("recv unexpected msg %d", recv.Msg.MsgId())
	}

	err = authAccept(Trans2Msg[*transport.MsgAuth](recv))
	if err != nil {
		return err
	}

	return nil
}

// ClientFinished 客户端握手结束
func (h *Handshake) ClientFinished(finishedAccept FinishedAccept) error {
	if finishedAccept == nil {
		return errors.New("finishedAccept is nil")
	}

	trans := Transceiver{
		Conn:       h.Conn,
		Encoder:    h.Encoder,
		Decoder:    h.Decoder,
		RetryTimes: h.RetryTimes,
	}

	defer trans.Decoder.GC()

	recv, err := trans.Recv()
	if err != nil {
		return err
	}

	switch recv.Msg.MsgId() {
	case transport.MsgId_Finished:
		break
	case transport.MsgId_Rst:
		return fmt.Errorf("recv rst, %s", Trans2Msg[*transport.MsgRst](recv).Msg.Message)
	default:
		return fmt.Errorf("recv unexpected msg %d", recv.Msg.MsgId())
	}

	err = finishedAccept(Trans2Msg[*transport.MsgFinished](recv))
	if err != nil {
		return err
	}

	return nil
}

// ServerFinished 服务端握手结束
func (h *Handshake) ServerFinished(finishedMsg Msg[*transport.MsgFinished]) error {
	trans := Transceiver{
		Conn:       h.Conn,
		Encoder:    h.Encoder,
		Decoder:    h.Decoder,
		RetryTimes: h.RetryTimes,
	}

	err := trans.Send(Msg2Trans(finishedMsg))
	if err != nil {
		return err
	}

	return nil
}
