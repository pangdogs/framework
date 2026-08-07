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
	// HelloAccept 在服务端校验客户端 Hello 并生成 Hello 响应。
	HelloAccept = generic.FuncPair1[Event[*gtp.MsgHello], Event[*gtp.MsgHello], error]
	// HelloFin 在客户端校验服务端 Hello 响应。
	HelloFin = generic.Func1[Event[*gtp.MsgHello], error]
	// SecretKeyExchangeAccept 在客户端处理服务端选择的密钥交换消息并生成响应。
	SecretKeyExchangeAccept = generic.FuncPair1[IEvent, IEvent, error]
	// ECDHESecretKeyExchangeFin 在服务端校验客户端 ECDHE 响应并生成密码规范切换消息。
	ECDHESecretKeyExchangeFin = generic.FuncPair1[Event[*gtp.MsgECDHESecretKeyExchange], Event[*gtp.MsgChangeCipherSpec], error]
	// ChangeCipherSpecAccept 在客户端校验服务端密码规范切换消息并生成响应。
	ChangeCipherSpecAccept = generic.FuncPair1[Event[*gtp.MsgChangeCipherSpec], Event[*gtp.MsgChangeCipherSpec], error]
	// ChangeCipherSpecFin 在服务端校验客户端密码规范切换响应。
	ChangeCipherSpecFin = generic.Func1[Event[*gtp.MsgChangeCipherSpec], error]
	// AuthAccept 在服务端校验客户端鉴权消息。
	AuthAccept = generic.Func1[Event[*gtp.MsgAuth], error]
	// ContinueAccept 在服务端校验并恢复客户端会话状态。
	ContinueAccept = generic.Func1[Event[*gtp.MsgContinue], error]
	// FinishedAccept 在客户端校验服务端握手完成消息。
	FinishedAccept = generic.Func1[Event[*gtp.MsgFinished], error]
)

// HandshakeProtocol 按客户端或服务端角色执行 GTP 握手阶段。
type HandshakeProtocol struct {
	Transceiver *Transceiver // 事件收发器。
	RetryTimes  int          // 网络 I/O 超时后的重试次数。
}

// ClientHello 发送客户端 Hello，接收并校验服务端 Hello。
func (h *HandshakeProtocol) ClientHello(ctx context.Context, hello Event[*gtp.MsgHello], helloFin HelloFin) (err error) {
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
		return fmt.Errorf("%w: %w", ErrProtocol, ToRstError(AssertEvent[*gtp.MsgRst](recv)))
	default:
		return fmt.Errorf("%w: %w (%d)", ErrProtocol, ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	err = helloFin.UnsafeCall(AssertEvent[*gtp.MsgHello](recv))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	return nil
}

// ServerHello 接收客户端 Hello，经回调生成并发送服务端 Hello；失败时尝试发送 RST。
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

	reply, err := helloAccept.UnsafeCall(AssertEvent[*gtp.MsgHello](recv))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	err = h.retrySend(trans.Send(reply.Interface()))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	return nil
}

// ClientSecretKeyExchange 完成客户端密钥交换和密码规范切换两个往返阶段。
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
	}()

	recv, err := h.retryRecv(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_ECDHESecretKeyExchange:
		break
	case gtp.MsgId_Rst:
		return fmt.Errorf("%w: %w", ErrProtocol, ToRstError(AssertEvent[*gtp.MsgRst](recv)))
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
		return fmt.Errorf("%w: %w", ErrProtocol, ToRstError(AssertEvent[*gtp.MsgRst](recv)))
	default:
		return fmt.Errorf("%w: %w (%d)", ErrProtocol, ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	changeCipherSpecReply, err := changeCipherSpecAccept.UnsafeCall(AssertEvent[*gtp.MsgChangeCipherSpec](recv))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	err = h.retrySend(trans.Send(changeCipherSpecReply.Interface()))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	return nil
}

// ServerECDHESecretKeyExchange 完成服务端 ECDHE 密钥交换和密码规范切换两个往返阶段。
func (h *HandshakeProtocol) ServerECDHESecretKeyExchange(ctx context.Context, secretKeyExchange Event[*gtp.MsgECDHESecretKeyExchange], secretKeyExchangeFin ECDHESecretKeyExchangeFin, changeCipherSpecFin ChangeCipherSpecFin) (err error) {
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

	changeCipherSpecMsg, err := secretKeyExchangeFin.UnsafeCall(AssertEvent[*gtp.MsgECDHESecretKeyExchange](recv))
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

	err = changeCipherSpecFin.UnsafeCall(AssertEvent[*gtp.MsgChangeCipherSpec](recv))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	return nil
}

// ClientAuth 发送客户端鉴权消息。
func (h *HandshakeProtocol) ClientAuth(ctx context.Context, auth Event[*gtp.MsgAuth]) (err error) {
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

// ServerAuth 接收并校验客户端鉴权消息；失败时尝试发送 RST。
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

	err = authAccept.UnsafeCall(AssertEvent[*gtp.MsgAuth](recv))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	return nil
}

// ClientContinue 发送客户端会话续接消息。
func (h *HandshakeProtocol) ClientContinue(ctx context.Context, cont Event[*gtp.MsgContinue]) (err error) {
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

// ServerContinue 接收并恢复客户端会话；失败时尝试发送 RST。
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

	err = continueAccept.UnsafeCall(AssertEvent[*gtp.MsgContinue](recv))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	return nil
}

// ClientFinished 接收并校验服务端握手完成消息。
func (h *HandshakeProtocol) ClientFinished(ctx context.Context, finishedAccept FinishedAccept) (err error) {
	if h.Transceiver == nil {
		return fmt.Errorf("%w: Transceiver is nil", ErrProtocol)
	}

	if finishedAccept == nil {
		return fmt.Errorf("%w: %w: finishedAccept is nil", ErrProtocol, core.ErrArgs)
	}

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w: %w", ErrProtocol, core.ErrPanicked, panicErr)
		}
	}()

	recv, err := h.retryRecv(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	switch recv.Msg.MsgId() {
	case gtp.MsgId_Finished:
		break
	case gtp.MsgId_Rst:
		return fmt.Errorf("%w: %w", ErrProtocol, ToRstError(AssertEvent[*gtp.MsgRst](recv)))
	default:
		return fmt.Errorf("%w: %w (%d)", ErrProtocol, ErrUnexpectedMsg, recv.Msg.MsgId())
	}

	err = finishedAccept.UnsafeCall(AssertEvent[*gtp.MsgFinished](recv))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	return nil
}

// ServerFinished 发送服务端握手完成消息；失败时尝试发送 RST。
func (h *HandshakeProtocol) ServerFinished(ctx context.Context, finished Event[*gtp.MsgFinished]) (err error) {
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
