package transport

import (
	"context"
	"errors"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/types"
	"git.golaxy.org/framework/net/gtp"
)

var (
	ErrUnexpectedMsg = errors.New("gtp: unexpected msg") // 收到非预期的消息
)

type (
	HelloAccept               = generic.PairFunc1[Event[gtp.MsgHello], Event[gtp.MsgHello], error]                             // 服务端确认客户端Hello请求
	HelloFin                  = generic.Func1[Event[gtp.MsgHello], error]                                                      // 客户端获取服务端Hello响应
	SecretKeyExchangeAccept   = generic.PairFunc1[Event[gtp.MsgReader], Event[gtp.MsgReader], error]                           // 客户端确认服务端SecretKeyExchange请求，需要自己判断消息Id并处理，用于支持多种秘钥交换函数
	ECDHESecretKeyExchangeFin = generic.PairFunc1[Event[gtp.MsgECDHESecretKeyExchange], Event[gtp.MsgChangeCipherSpec], error] // 服务端获取客户端ECDHESecretKeyExchange响应
	ChangeCipherSpecAccept    = generic.PairFunc1[Event[gtp.MsgChangeCipherSpec], Event[gtp.MsgChangeCipherSpec], error]       // 客户端确认服务端ChangeCipherSpec请求
	ChangeCipherSpecFin       = generic.Func1[Event[gtp.MsgChangeCipherSpec], error]                                           // 服务端获取客户端ChangeCipherSpec响应
	AuthAccept                = generic.Func1[Event[gtp.MsgAuth], error]                                                       // 服务端确认客户端Auth请求
	ContinueAccept            = generic.Func1[Event[gtp.MsgContinue], error]                                                   // 服务端确认客户端Continue请求
	FinishedAccept            = generic.Func1[Event[gtp.MsgFinished], error]                                                   // 客户端确认服务端Finished请求
)

// HandshakeProtocol 握手协议
type HandshakeProtocol struct {
	Transceiver *Transceiver // 消息事件收发器
	RetryTimes  int          // 网络io超时时的重试次数
}

// ClientHello 客户端Hello
func (h *HandshakeProtocol) ClientHello(ctx context.Context, hello Event[gtp.MsgHello], helloFin HelloFin) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if helloFin == nil {
		return fmt.Errorf("%w: helloFin is nil", core.ErrArgs)
	}

	if h.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	trans := h.Transceiver

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
		trans.GC()
	}()

	err = h.retrySend(trans.Send(hello.Interface()))
	if err != nil {
		return err
	}

	recv, err := h.retryRecv(ctx)
	if err != nil {
		return err
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_Hello:
		break
	case gtp.MsgId_Rst:
		return CastRstErr(EventT[gtp.MsgRst](recv))
	default:
		return fmt.Errorf("%w (%d)", ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	err = helloFin.Exec(EventT[gtp.MsgHello](recv))
	if err != nil {
		return err
	}

	return nil
}

// ServerHello 服务端Hello
func (h *HandshakeProtocol) ServerHello(ctx context.Context, helloAccept HelloAccept) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if helloAccept == nil {
		return fmt.Errorf("%w: helloAccept is nil", core.ErrArgs)
	}

	if h.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	trans := h.Transceiver

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
		if err != nil {
			trans.SendRst(err)
		}
		trans.GC()
	}()

	recv, err := h.retryRecv(ctx)
	if err != nil {
		return err
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_Hello:
		break
	default:
		return fmt.Errorf("%w (%d)", ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	reply, err := helloAccept.Exec(EventT[gtp.MsgHello](recv))
	if err != nil {
		return err
	}

	err = h.retrySend(trans.Send(reply.Interface()))
	if err != nil {
		return err
	}

	return nil
}

// ClientSecretKeyExchange 客户端交换秘钥
func (h *HandshakeProtocol) ClientSecretKeyExchange(ctx context.Context, secretKeyExchangeAccept SecretKeyExchangeAccept, changeCipherSpecAccept ChangeCipherSpecAccept) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if secretKeyExchangeAccept == nil {
		return fmt.Errorf("%w: secretKeyExchangeAccept is nil", core.ErrArgs)
	}

	if changeCipherSpecAccept == nil {
		return fmt.Errorf("%w: changeCipherSpecAccept is nil", core.ErrArgs)
	}

	if h.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	trans := h.Transceiver

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
		trans.GC()
	}()

	recv, err := h.retryRecv(ctx)
	if err != nil {
		return err
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_ECDHESecretKeyExchange:
		break
	case gtp.MsgId_Rst:
		return CastRstErr(EventT[gtp.MsgRst](recv))
	default:
		return fmt.Errorf("%w (%d)", ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	secretKeyExchangeReply, err := secretKeyExchangeAccept.Exec(recv)
	if err != nil {
		return err
	}

	err = h.retrySend(trans.Send(secretKeyExchangeReply.Interface()))
	if err != nil {
		return err
	}

	recv, err = h.retryRecv(ctx)
	if err != nil {
		return err
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_ChangeCipherSpec:
		break
	case gtp.MsgId_Rst:
		return CastRstErr(EventT[gtp.MsgRst](recv))
	default:
		return fmt.Errorf("%w (%d)", ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	changeCipherSpecReply, err := changeCipherSpecAccept.Exec(EventT[gtp.MsgChangeCipherSpec](recv))
	if err != nil {
		return err
	}

	err = h.retrySend(trans.Send(changeCipherSpecReply.Interface()))
	if err != nil {
		return err
	}

	return nil
}

// ServerECDHESecretKeyExchange 服务端交换秘钥（ECDHE）
func (h *HandshakeProtocol) ServerECDHESecretKeyExchange(ctx context.Context, secretKeyExchange Event[gtp.MsgECDHESecretKeyExchange], secretKeyExchangeFin ECDHESecretKeyExchangeFin, changeCipherSpecFin ChangeCipherSpecFin) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if secretKeyExchangeFin == nil {
		return fmt.Errorf("%w: secretKeyExchangeFin is nil", core.ErrArgs)
	}

	if changeCipherSpecFin == nil {
		return fmt.Errorf("%w: changeCipherSpecFin is nil", core.ErrArgs)
	}

	if h.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	trans := h.Transceiver

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
		if err != nil {
			trans.SendRst(err)
		}
		trans.GC()
	}()

	err = h.retrySend(trans.Send(secretKeyExchange.Interface()))
	if err != nil {
		return err
	}

	recv, err := h.retryRecv(ctx)
	if err != nil {
		return err
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_ECDHESecretKeyExchange:
		break
	default:
		return fmt.Errorf("%w (%d)", ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	changeCipherSpecMsg, err := secretKeyExchangeFin.Exec(EventT[gtp.MsgECDHESecretKeyExchange](recv))
	if err != nil {
		return err
	}

	err = h.retrySend(trans.Send(changeCipherSpecMsg.Interface()))
	if err != nil {
		return err
	}

	recv, err = h.retryRecv(ctx)
	if err != nil {
		return err
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_ChangeCipherSpec:
		break
	default:
		return fmt.Errorf("%w (%d)", ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	err = changeCipherSpecFin.Exec(EventT[gtp.MsgChangeCipherSpec](recv))
	if err != nil {
		return err
	}

	return nil
}

// ClientAuth 客户端发起鉴权
func (h *HandshakeProtocol) ClientAuth(ctx context.Context, auth Event[gtp.MsgAuth]) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if h.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	trans := h.Transceiver

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
	}()

	err = h.retrySend(trans.Send(auth.Interface()))
	if err != nil {
		return err
	}

	return nil
}

// ServerAuth 服务端验证鉴权
func (h *HandshakeProtocol) ServerAuth(ctx context.Context, authAccept AuthAccept) (err error) {
	if authAccept == nil {
		return fmt.Errorf("%w: authAccept is nil", core.ErrArgs)
	}

	if h.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	trans := h.Transceiver

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
		if err != nil {
			trans.SendRst(err)
		}
		trans.GC()
	}()

	recv, err := h.retryRecv(ctx)
	if err != nil {
		return err
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_Auth:
		break
	default:
		return fmt.Errorf("%w (%d)", ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	err = authAccept.Exec(EventT[gtp.MsgAuth](recv))
	if err != nil {
		return err
	}

	return nil
}

// ClientContinue 客户端发起重连
func (h *HandshakeProtocol) ClientContinue(ctx context.Context, cont Event[gtp.MsgContinue]) (err error) {
	if h.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	trans := h.Transceiver

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
	}()

	err = h.retrySend(trans.Send(cont.Interface()))
	if err != nil {
		return err
	}

	return nil
}

// ServerContinue 服务端处理重连
func (h *HandshakeProtocol) ServerContinue(ctx context.Context, continueAccept ContinueAccept) (err error) {
	if continueAccept == nil {
		return fmt.Errorf("%w: continueAccept is nil", core.ErrArgs)
	}

	if h.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	trans := h.Transceiver

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
		if err != nil {
			trans.SendRst(err)
		}
		trans.GC()
	}()

	recv, err := h.retryRecv(ctx)
	if err != nil {
		return err
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_Continue:
		break
	default:
		return fmt.Errorf("%w (%d)", ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	err = continueAccept.Exec(EventT[gtp.MsgContinue](recv))
	if err != nil {
		return err
	}

	return nil
}

// ClientFinished 客户端握手结束
func (h *HandshakeProtocol) ClientFinished(ctx context.Context, finishedAccept FinishedAccept) (err error) {
	if finishedAccept == nil {
		return fmt.Errorf("%w: finishedAccept is nil", core.ErrArgs)
	}

	if h.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	trans := h.Transceiver

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
		trans.GC()
	}()

	recv, err := h.retryRecv(ctx)
	if err != nil {
		return err
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_Finished:
		break
	case gtp.MsgId_Rst:
		return CastRstErr(EventT[gtp.MsgRst](recv))
	default:
		return fmt.Errorf("%w (%d)", ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	err = finishedAccept.Exec(EventT[gtp.MsgFinished](recv))
	if err != nil {
		return err
	}

	return nil
}

// ServerFinished 服务端握手结束
func (h *HandshakeProtocol) ServerFinished(ctx context.Context, finished Event[gtp.MsgFinished]) (err error) {
	if h.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	trans := h.Transceiver

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
		if err != nil {
			trans.SendRst(err)
		}
	}()

	err = h.retrySend(trans.Send(finished.Interface()))
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

func (h *HandshakeProtocol) retryRecv(ctx context.Context) (Event[gtp.MsgReader], error) {
	e, err := h.Transceiver.Recv(ctx)
	return Retry{
		Transceiver: h.Transceiver,
		Times:       h.RetryTimes,
		Ctx:         ctx,
	}.Recv(e, err)
}
