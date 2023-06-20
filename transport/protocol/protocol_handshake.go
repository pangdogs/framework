package tcp

import (
	"errors"
	"fmt"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/codec"
	"net"
)

type (
	HelloAccept               = func(Event[*transport.MsgHello]) (Event[*transport.MsgHello], error)
	HelloFin                  = func(Event[*transport.MsgHello]) error
	SecretKeyExchangeAccept   = func(Event[transport.Msg]) (Event[transport.Msg], error)
	ECDHESecretKeyExchangeFin = func(Event[*transport.MsgECDHESecretKeyExchange]) (Event[*transport.MsgChangeCipherSpec], error)
	ChangeCipherSpecAccept    = func(Event[*transport.MsgChangeCipherSpec]) (Event[*transport.MsgChangeCipherSpec], error)
	ChangeCipherSpecFin       = func(Event[*transport.MsgChangeCipherSpec]) error
	AuthAccept                = func(Event[*transport.MsgAuth]) error
	FinishedAccept            = func(Event[*transport.MsgFinished]) error
)

// Handshake 握手协议
type Handshake struct {
	Conn       net.Conn       // 网络连接
	Encoder    codec.IEncoder // 消息包编码器
	Decoder    codec.IDecoder // 消息包解码器
	RetryTimes int            // io超时重试次数
}

// ClientHello 客户端Hello
func (h *Handshake) ClientHello(hello Event[*transport.MsgHello], helloFin HelloFin) error {
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

	err := trans.Send(PackEvent(hello))
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
		return EventToRstErr(UnpackEvent[*transport.MsgRst](recv))
	default:
		return fmt.Errorf("recv unexpected msg %d", recv.Msg.MsgId())
	}

	err = helloFin(UnpackEvent[*transport.MsgHello](recv))
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

	reply, err := helloAccept(UnpackEvent[*transport.MsgHello](recv))
	if err != nil {
		return err
	}

	err = trans.Send(PackEvent(reply))
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
		return EventToRstErr(UnpackEvent[*transport.MsgRst](recv))
	default:
		return fmt.Errorf("recv unexpected msg %d", recv.Msg.MsgId())
	}

	secretKeyExchangeReply, err := secretKeyExchangeAccept(recv)
	if err != nil {
		return err
	}

	err = trans.Send(PackEvent(secretKeyExchangeReply))
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
		return EventToRstErr(UnpackEvent[*transport.MsgRst](recv))
	default:
		return fmt.Errorf("recv unexpected msg %d", recv.Msg.MsgId())
	}

	changeCipherSpecReply, err := changeCipherSpecAccept(UnpackEvent[*transport.MsgChangeCipherSpec](recv))
	if err != nil {
		return err
	}

	err = trans.Send(PackEvent(changeCipherSpecReply))
	if err != nil {
		return err
	}

	return nil
}

// ServerECDHESecretKeyExchange 服务端交换秘钥（ECDHE）
func (h *Handshake) ServerECDHESecretKeyExchange(secretKeyExchange Event[*transport.MsgECDHESecretKeyExchange], secretKeyExchangeFin ECDHESecretKeyExchangeFin, changeCipherSpecFin ChangeCipherSpecFin) (err error) {
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

	err = trans.Send(PackEvent(secretKeyExchange))
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

	changeCipherSpecMsg, err := secretKeyExchangeFin(UnpackEvent[*transport.MsgECDHESecretKeyExchange](recv))
	if err != nil {
		return err
	}

	err = trans.Send(PackEvent(changeCipherSpecMsg))
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

	err = changeCipherSpecFin(UnpackEvent[*transport.MsgChangeCipherSpec](recv))
	if err != nil {
		return err
	}

	return nil
}

// ClientAuth 客户端发起鉴权
func (h *Handshake) ClientAuth(auth Event[*transport.MsgAuth]) error {
	trans := Transceiver{
		Conn:       h.Conn,
		Encoder:    h.Encoder,
		Decoder:    h.Decoder,
		RetryTimes: h.RetryTimes,
	}

	err := trans.Send(PackEvent(auth))
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

	err = authAccept(UnpackEvent[*transport.MsgAuth](recv))
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
		return EventToRstErr(UnpackEvent[*transport.MsgRst](recv))
	default:
		return fmt.Errorf("recv unexpected msg %d", recv.Msg.MsgId())
	}

	err = finishedAccept(UnpackEvent[*transport.MsgFinished](recv))
	if err != nil {
		return err
	}

	return nil
}

// ServerFinished 服务端握手结束
func (h *Handshake) ServerFinished(finished Event[*transport.MsgFinished]) error {
	trans := Transceiver{
		Conn:       h.Conn,
		Encoder:    h.Encoder,
		Decoder:    h.Decoder,
		RetryTimes: h.RetryTimes,
	}

	err := trans.Send(PackEvent(finished))
	if err != nil {
		return err
	}

	return nil
}
