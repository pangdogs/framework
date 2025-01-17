/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package transport

import (
	"context"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/types"
	"git.golaxy.org/framework/net/gtp"
)

type (
	HelloAccept               = generic.FuncPair1[Event[gtp.MsgHello], Event[gtp.MsgHello], error]                             // 服务端确认客户端Hello请求
	HelloFin                  = generic.Func1[Event[gtp.MsgHello], error]                                                      // 客户端获取服务端Hello响应
	SecretKeyExchangeAccept   = generic.FuncPair1[IEvent, IEvent, error]                                                       // 客户端确认服务端SecretKeyExchange请求，需要自己判断消息Id并处理，用于支持多种秘钥交换函数
	ECDHESecretKeyExchangeFin = generic.FuncPair1[Event[gtp.MsgECDHESecretKeyExchange], Event[gtp.MsgChangeCipherSpec], error] // 服务端获取客户端ECDHESecretKeyExchange响应
	ChangeCipherSpecAccept    = generic.FuncPair1[Event[gtp.MsgChangeCipherSpec], Event[gtp.MsgChangeCipherSpec], error]       // 客户端确认服务端ChangeCipherSpec请求
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
	if h.Transceiver == nil {
		return fmt.Errorf("%w: Transceiver is nil", ErrProtocol)
	}
	trans := h.Transceiver

	if ctx == nil {
		ctx = context.Background()
	}

	if helloFin == nil {
		return fmt.Errorf("%w: %w: helloFin is nil", ErrProtocol, core.ErrArgs)
	}

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w: %w", ErrProtocol, core.ErrPanicked, panicErr)
		}
		trans.GC()
	}()

	err = h.retrySend(trans.Send(hello.Interface()))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	recv, err := h.retryRecv(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_Hello:
		break
	case gtp.MsgId_Rst:
		return fmt.Errorf("%w: %w", ErrProtocol, CastRstErr(EventT[gtp.MsgRst](recv)))
	default:
		return fmt.Errorf("%w: %w (%d)", ErrProtocol, ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	err = helloFin.UnsafeCall(EventT[gtp.MsgHello](recv))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	return nil
}

// ServerHello 服务端Hello
func (h *HandshakeProtocol) ServerHello(ctx context.Context, helloAccept HelloAccept) (err error) {
	if h.Transceiver == nil {
		return fmt.Errorf("%w: Transceiver is nil", ErrProtocol)
	}
	trans := h.Transceiver

	if ctx == nil {
		ctx = context.Background()
	}

	if helloAccept == nil {
		return fmt.Errorf("%w: %w: helloAccept is nil", ErrProtocol, core.ErrArgs)
	}

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w: %w", ErrProtocol, core.ErrPanicked, panicErr)
		}
		if err != nil {
			trans.SendRst(err)
		}
		trans.GC()
	}()

	recv, err := h.retryRecv(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_Hello:
		break
	default:
		return fmt.Errorf("%w: %w (%d)", ErrProtocol, ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	reply, err := helloAccept.UnsafeCall(EventT[gtp.MsgHello](recv))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	err = h.retrySend(trans.Send(reply.Interface()))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	return nil
}

// ClientSecretKeyExchange 客户端交换秘钥
func (h *HandshakeProtocol) ClientSecretKeyExchange(ctx context.Context, secretKeyExchangeAccept SecretKeyExchangeAccept, changeCipherSpecAccept ChangeCipherSpecAccept) (err error) {
	if h.Transceiver == nil {
		return fmt.Errorf("%w: Transceiver is nil", ErrProtocol)
	}
	trans := h.Transceiver

	if ctx == nil {
		ctx = context.Background()
	}

	if secretKeyExchangeAccept == nil {
		return fmt.Errorf("%w: %w: secretKeyExchangeAccept is nil", ErrProtocol, core.ErrArgs)
	}

	if changeCipherSpecAccept == nil {
		return fmt.Errorf("%w: %w: changeCipherSpecAccept is nil", ErrProtocol, core.ErrArgs)
	}

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w: %w", ErrProtocol, core.ErrPanicked, panicErr)
		}
		trans.GC()
	}()

	recv, err := h.retryRecv(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_ECDHESecretKeyExchange:
		break
	case gtp.MsgId_Rst:
		return fmt.Errorf("%w: %w", ErrProtocol, CastRstErr(EventT[gtp.MsgRst](recv)))
	default:
		return fmt.Errorf("%w: %w (%d)", ErrProtocol, ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	secretKeyExchangeReply, err := secretKeyExchangeAccept.UnsafeCall(recv)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	err = h.retrySend(trans.Send(secretKeyExchangeReply.Interface()))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	recv, err = h.retryRecv(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_ChangeCipherSpec:
		break
	case gtp.MsgId_Rst:
		return fmt.Errorf("%w: %w", ErrProtocol, CastRstErr(EventT[gtp.MsgRst](recv)))
	default:
		return fmt.Errorf("%w: %w (%d)", ErrProtocol, ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	changeCipherSpecReply, err := changeCipherSpecAccept.UnsafeCall(EventT[gtp.MsgChangeCipherSpec](recv))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	err = h.retrySend(trans.Send(changeCipherSpecReply.Interface()))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	return nil
}

// ServerECDHESecretKeyExchange 服务端交换秘钥（ECDHE）
func (h *HandshakeProtocol) ServerECDHESecretKeyExchange(ctx context.Context, secretKeyExchange Event[gtp.MsgECDHESecretKeyExchange], secretKeyExchangeFin ECDHESecretKeyExchangeFin, changeCipherSpecFin ChangeCipherSpecFin) (err error) {
	if h.Transceiver == nil {
		return fmt.Errorf("%w: Transceiver is nil", ErrProtocol)
	}
	trans := h.Transceiver

	if ctx == nil {
		ctx = context.Background()
	}

	if secretKeyExchangeFin == nil {
		return fmt.Errorf("%w: %w: secretKeyExchangeFin is nil", ErrProtocol, core.ErrArgs)
	}

	if changeCipherSpecFin == nil {
		return fmt.Errorf("%w: %w: changeCipherSpecFin is nil", ErrProtocol, core.ErrArgs)
	}

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w: %w", ErrProtocol, core.ErrPanicked, panicErr)
		}
		if err != nil {
			trans.SendRst(err)
		}
		trans.GC()
	}()

	err = h.retrySend(trans.Send(secretKeyExchange.Interface()))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	recv, err := h.retryRecv(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_ECDHESecretKeyExchange:
		break
	default:
		return fmt.Errorf("%w: %w (%d)", ErrProtocol, ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	changeCipherSpecMsg, err := secretKeyExchangeFin.UnsafeCall(EventT[gtp.MsgECDHESecretKeyExchange](recv))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	err = h.retrySend(trans.Send(changeCipherSpecMsg.Interface()))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	recv, err = h.retryRecv(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_ChangeCipherSpec:
		break
	default:
		return fmt.Errorf("%w: %w (%d)", ErrProtocol, ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	err = changeCipherSpecFin.UnsafeCall(EventT[gtp.MsgChangeCipherSpec](recv))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	return nil
}

// ClientAuth 客户端发起鉴权
func (h *HandshakeProtocol) ClientAuth(ctx context.Context, auth Event[gtp.MsgAuth]) (err error) {
	if h.Transceiver == nil {
		return fmt.Errorf("%w: Transceiver is nil", ErrProtocol)
	}
	trans := h.Transceiver

	if ctx == nil {
		ctx = context.Background()
	}

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w: %w", ErrProtocol, core.ErrPanicked, panicErr)
		}
	}()

	err = h.retrySend(trans.Send(auth.Interface()))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	return nil
}

// ServerAuth 服务端验证鉴权
func (h *HandshakeProtocol) ServerAuth(ctx context.Context, authAccept AuthAccept) (err error) {
	if h.Transceiver == nil {
		return fmt.Errorf("%w: Transceiver is nil", ErrProtocol)
	}
	trans := h.Transceiver

	if authAccept == nil {
		return fmt.Errorf("%w: %w: authAccept is nil", ErrProtocol, core.ErrArgs)
	}

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w: %w", ErrProtocol, core.ErrPanicked, panicErr)
		}
		if err != nil {
			trans.SendRst(err)
		}
		trans.GC()
	}()

	recv, err := h.retryRecv(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_Auth:
		break
	default:
		return fmt.Errorf("%w: %w (%d)", ErrProtocol, ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	err = authAccept.UnsafeCall(EventT[gtp.MsgAuth](recv))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	return nil
}

// ClientContinue 客户端发起重连
func (h *HandshakeProtocol) ClientContinue(ctx context.Context, cont Event[gtp.MsgContinue]) (err error) {
	if h.Transceiver == nil {
		return fmt.Errorf("%w: Transceiver is nil", ErrProtocol)
	}
	trans := h.Transceiver

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w: %w", ErrProtocol, core.ErrPanicked, panicErr)
		}
	}()

	err = h.retrySend(trans.Send(cont.Interface()))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	return nil
}

// ServerContinue 服务端处理重连
func (h *HandshakeProtocol) ServerContinue(ctx context.Context, continueAccept ContinueAccept) (err error) {
	if h.Transceiver == nil {
		return fmt.Errorf("%w: Transceiver is nil", ErrProtocol)
	}
	trans := h.Transceiver

	if continueAccept == nil {
		return fmt.Errorf("%w: %w: continueAccept is nil", ErrProtocol, core.ErrArgs)
	}

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w: %w", ErrProtocol, core.ErrPanicked, panicErr)
		}
		if err != nil {
			trans.SendRst(err)
		}
		trans.GC()
	}()

	recv, err := h.retryRecv(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_Continue:
		break
	default:
		return fmt.Errorf("%w: %w (%d)", ErrProtocol, ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	err = continueAccept.UnsafeCall(EventT[gtp.MsgContinue](recv))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	return nil
}

// ClientFinished 客户端握手结束
func (h *HandshakeProtocol) ClientFinished(ctx context.Context, finishedAccept FinishedAccept) (err error) {
	if h.Transceiver == nil {
		return fmt.Errorf("%w: Transceiver is nil", ErrProtocol)
	}
	trans := h.Transceiver

	if finishedAccept == nil {
		return fmt.Errorf("%w: %w: finishedAccept is nil", ErrProtocol, core.ErrArgs)
	}

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w: %w", ErrProtocol, core.ErrPanicked, panicErr)
		}
		trans.GC()
	}()

	recv, err := h.retryRecv(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_Finished:
		break
	case gtp.MsgId_Rst:
		return fmt.Errorf("%w: %w", ErrProtocol, CastRstErr(EventT[gtp.MsgRst](recv)))
	default:
		return fmt.Errorf("%w: %w (%d)", ErrProtocol, ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	err = finishedAccept.UnsafeCall(EventT[gtp.MsgFinished](recv))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	return nil
}

// ServerFinished 服务端握手结束
func (h *HandshakeProtocol) ServerFinished(ctx context.Context, finished Event[gtp.MsgFinished]) (err error) {
	if h.Transceiver == nil {
		return fmt.Errorf("%w: Transceiver is nil", ErrProtocol)
	}
	trans := h.Transceiver

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w: %w", ErrProtocol, core.ErrPanicked, panicErr)
		}
		if err != nil {
			trans.SendRst(err)
		}
	}()

	err = h.retrySend(trans.Send(finished.Interface()))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	return nil
}

func (h *HandshakeProtocol) retrySend(err error) error {
	return Retry{
		Transceiver: h.Transceiver,
		Times:       h.RetryTimes,
	}.Send(err)
}

func (h *HandshakeProtocol) retryRecv(ctx context.Context) (IEvent, error) {
	e, err := h.Transceiver.Recv(ctx)
	return Retry{
		Transceiver: h.Transceiver,
		Times:       h.RetryTimes,
		Ctx:         ctx,
	}.Recv(e, err)
}
