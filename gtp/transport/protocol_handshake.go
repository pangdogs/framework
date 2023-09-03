package transport

import (
	"errors"
	"fmt"
	"kit.golaxy.org/golaxy/util"
	"kit.golaxy.org/plugins/gtp"
)

type (
	HelloAccept               = func(Event[*gtp.MsgHello]) (Event[*gtp.MsgHello], error)                             // 服务端确认客户端Hello请求
	HelloFin                  = func(Event[*gtp.MsgHello]) error                                                     // 客户端获取服务端Hello响应
	SecretKeyExchangeAccept   = func(Event[gtp.Msg]) (Event[gtp.Msg], error)                                         // 客户端确认服务端SecretKeyExchange请求，需要自己判断消息Id并处理，用于支持多种秘钥交换函数
	ECDHESecretKeyExchangeFin = func(Event[*gtp.MsgECDHESecretKeyExchange]) (Event[*gtp.MsgChangeCipherSpec], error) // 服务端获取客户端ECDHESecretKeyExchange响应
	ChangeCipherSpecAccept    = func(Event[*gtp.MsgChangeCipherSpec]) (Event[*gtp.MsgChangeCipherSpec], error)       // 客户端确认服务端ChangeCipherSpec请求
	ChangeCipherSpecFin       = func(Event[*gtp.MsgChangeCipherSpec]) error                                          // 服务端获取客户端ChangeCipherSpec响应
	AuthAccept                = func(Event[*gtp.MsgAuth]) error                                                      // 服务端确认客户端Auth请求
	ContinueAccept            = func(Event[*gtp.MsgContinue]) error                                                  // 服务端确认客户端Continue请求
	FinishedAccept            = func(Event[*gtp.MsgFinished]) error                                                  // 客户端确认服务端Finished请求
)

// HandshakeProtocol 握手协议
type HandshakeProtocol struct {
	Transceiver *Transceiver // 消息事件收发器
	RetryTimes  int          // 网络io超时时的重试次数
}

// ClientHello 客户端Hello
func (h *HandshakeProtocol) ClientHello(hello Event[*gtp.MsgHello], helloFin HelloFin) (err error) {
	if helloFin == nil {
		return errors.New("helloFin is nil")
	}

	if h.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	trans := h.Transceiver

	defer func() {
		if panicErr := util.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("panicked: %w", panicErr)
		}
		trans.GC()
	}()

	err = h.retrySend(trans.Send(PackEvent(hello)))
	if err != nil {
		return err
	}

	recv, err := h.retryRecv(trans.Recv())
	if err != nil {
		return err
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_Hello:
		break
	case gtp.MsgId_Rst:
		return EventToRstErr(UnpackEvent[*gtp.MsgRst](recv))
	default:
		return fmt.Errorf("%w: %d", ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	err = helloFin(UnpackEvent[*gtp.MsgHello](recv))
	if err != nil {
		return err
	}

	return nil
}

// ServerHello 服务端Hello
func (h *HandshakeProtocol) ServerHello(helloAccept HelloAccept) (err error) {
	if helloAccept == nil {
		return errors.New("helloAccept is nil")
	}

	if h.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	trans := h.Transceiver

	defer func() {
		if panicErr := util.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("panicked: %w", panicErr)
		}
		if err != nil {
			trans.SendRst(err)
		}
		trans.GC()
	}()

	recv, err := h.retryRecv(trans.Recv())
	if err != nil {
		return err
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_Hello:
		break
	default:
		return fmt.Errorf("%w: %d", ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	reply, err := helloAccept(UnpackEvent[*gtp.MsgHello](recv))
	if err != nil {
		return err
	}

	err = h.retrySend(trans.Send(PackEvent(reply)))
	if err != nil {
		return err
	}

	return nil
}

// ClientSecretKeyExchange 客户端交换秘钥
func (h *HandshakeProtocol) ClientSecretKeyExchange(secretKeyExchangeAccept SecretKeyExchangeAccept, changeCipherSpecAccept ChangeCipherSpecAccept) (err error) {
	if secretKeyExchangeAccept == nil {
		return errors.New("secretKeyExchangeAccept is nil")
	}

	if changeCipherSpecAccept == nil {
		return errors.New("changeCipherSpecAccept is nil")
	}

	if h.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	trans := h.Transceiver

	defer func() {
		if panicErr := util.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("panicked: %w", panicErr)
		}
		trans.GC()
	}()

	recv, err := h.retryRecv(trans.Recv())
	if err != nil {
		return err
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_ECDHESecretKeyExchange:
		break
	case gtp.MsgId_Rst:
		return EventToRstErr(UnpackEvent[*gtp.MsgRst](recv))
	default:
		return fmt.Errorf("%w: %d", ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	secretKeyExchangeReply, err := secretKeyExchangeAccept(recv)
	if err != nil {
		return err
	}

	err = h.retrySend(trans.Send(PackEvent(secretKeyExchangeReply)))
	if err != nil {
		return err
	}

	recv, err = h.retryRecv(trans.Recv())
	if err != nil {
		return err
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_ChangeCipherSpec:
		break
	case gtp.MsgId_Rst:
		return EventToRstErr(UnpackEvent[*gtp.MsgRst](recv))
	default:
		return fmt.Errorf("%w: %d", ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	changeCipherSpecReply, err := changeCipherSpecAccept(UnpackEvent[*gtp.MsgChangeCipherSpec](recv))
	if err != nil {
		return err
	}

	err = h.retrySend(trans.Send(PackEvent(changeCipherSpecReply)))
	if err != nil {
		return err
	}

	return nil
}

// ServerECDHESecretKeyExchange 服务端交换秘钥（ECDHE）
func (h *HandshakeProtocol) ServerECDHESecretKeyExchange(secretKeyExchange Event[*gtp.MsgECDHESecretKeyExchange], secretKeyExchangeFin ECDHESecretKeyExchangeFin, changeCipherSpecFin ChangeCipherSpecFin) (err error) {
	if secretKeyExchangeFin == nil {
		return errors.New("secretKeyExchangeFin is nil")
	}

	if changeCipherSpecFin == nil {
		return errors.New("changeCipherSpecFin is nil")
	}

	if h.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	trans := h.Transceiver

	defer func() {
		if panicErr := util.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("panicked: %w", panicErr)
		}
		if err != nil {
			trans.SendRst(err)
		}
		trans.GC()
	}()

	err = h.retrySend(trans.Send(PackEvent(secretKeyExchange)))
	if err != nil {
		return err
	}

	recv, err := h.retryRecv(trans.Recv())
	if err != nil {
		return err
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_ECDHESecretKeyExchange:
		break
	default:
		return fmt.Errorf("%w: %d", ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	changeCipherSpecMsg, err := secretKeyExchangeFin(UnpackEvent[*gtp.MsgECDHESecretKeyExchange](recv))
	if err != nil {
		return err
	}

	err = h.retrySend(trans.Send(PackEvent(changeCipherSpecMsg)))
	if err != nil {
		return err
	}

	recv, err = h.retryRecv(trans.Recv())
	if err != nil {
		return err
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_ChangeCipherSpec:
		break
	default:
		return fmt.Errorf("%w: %d", ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	err = changeCipherSpecFin(UnpackEvent[*gtp.MsgChangeCipherSpec](recv))
	if err != nil {
		return err
	}

	return nil
}

// ClientAuth 客户端发起鉴权
func (h *HandshakeProtocol) ClientAuth(auth Event[*gtp.MsgAuth]) (err error) {
	if h.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	trans := h.Transceiver

	defer func() {
		if panicErr := util.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("panicked: %w", panicErr)
		}
	}()

	err = h.retrySend(trans.Send(PackEvent(auth)))
	if err != nil {
		return err
	}

	return nil
}

// ServerAuth 服务端验证鉴权
func (h *HandshakeProtocol) ServerAuth(authAccept AuthAccept) (err error) {
	if authAccept == nil {
		return errors.New("authAccept is nil")
	}

	if h.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	trans := h.Transceiver

	defer func() {
		if panicErr := util.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("panicked: %w", panicErr)
		}
		if err != nil {
			trans.SendRst(err)
		}
		trans.GC()
	}()

	recv, err := h.retryRecv(trans.Recv())
	if err != nil {
		return err
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_Auth:
		break
	default:
		return fmt.Errorf("%w: %d", ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	err = authAccept(UnpackEvent[*gtp.MsgAuth](recv))
	if err != nil {
		return err
	}

	return nil
}

// ClientContinue 客户端发起重连
func (h *HandshakeProtocol) ClientContinue(cont Event[*gtp.MsgContinue]) (err error) {
	if h.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	trans := h.Transceiver

	defer func() {
		if panicErr := util.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("panicked: %w", panicErr)
		}
	}()

	err = h.retrySend(trans.Send(PackEvent(cont)))
	if err != nil {
		return err
	}

	return nil
}

// ServerContinue 服务端处理重连
func (h *HandshakeProtocol) ServerContinue(continueAccept ContinueAccept) (err error) {
	if continueAccept == nil {
		return errors.New("continueAccept is nil")
	}

	if h.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	trans := h.Transceiver

	defer func() {
		if panicErr := util.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("panicked: %w", panicErr)
		}
		if err != nil {
			trans.SendRst(err)
		}
		trans.GC()
	}()

	recv, err := h.retryRecv(trans.Recv())
	if err != nil {
		return err
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_Continue:
		break
	default:
		return fmt.Errorf("%w: %d", ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	err = continueAccept(UnpackEvent[*gtp.MsgContinue](recv))
	if err != nil {
		return err
	}

	return nil
}

// ClientFinished 客户端握手结束
func (h *HandshakeProtocol) ClientFinished(finishedAccept FinishedAccept) (err error) {
	if finishedAccept == nil {
		return errors.New("finishedAccept is nil")
	}

	if h.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	trans := h.Transceiver

	defer func() {
		if panicErr := util.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("panicked: %w", panicErr)
		}
		trans.GC()
	}()

	recv, err := h.retryRecv(trans.Recv())
	if err != nil {
		return err
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_Finished:
		break
	case gtp.MsgId_Rst:
		return EventToRstErr(UnpackEvent[*gtp.MsgRst](recv))
	default:
		return fmt.Errorf("%w: %d", ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	err = finishedAccept(UnpackEvent[*gtp.MsgFinished](recv))
	if err != nil {
		return err
	}

	return nil
}

// ServerFinished 服务端握手结束
func (h *HandshakeProtocol) ServerFinished(finished Event[*gtp.MsgFinished]) (err error) {
	if h.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	trans := h.Transceiver

	defer func() {
		if panicErr := util.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("panicked: %w", panicErr)
		}
		if err != nil {
			trans.SendRst(err)
		}
	}()

	err = h.retrySend(trans.Send(PackEvent(finished)))
	if err != nil {
		return err
	}

	return nil
}

func (h *HandshakeProtocol) retrySend(err error) error {
	return Retry{
		Transceiver: h.Transceiver,
		Times:       h.RetryTimes,
	}.Send(err)
}

func (h *HandshakeProtocol) retryRecv(e Event[gtp.Msg], err error) (Event[gtp.Msg], error) {
	return Retry{
		Transceiver: h.Transceiver,
		Times:       h.RetryTimes,
	}.Recv(e, err)
}