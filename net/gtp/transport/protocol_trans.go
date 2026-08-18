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
	"fmt"

	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/net/gtp"
)

type (
	// PayloadHandler 处理业务负载事件。
	PayloadHandler = generic.DelegateVoid1[Event[*gtp.MsgPayload]]
)

// TransProtocol 发送业务负载并将收到的负载事件交给处理器。
type TransProtocol struct {
	AutoRecover    bool           // 是否恢复负载处理器的 panic。
	ReportError    chan error     // 恢复 panic 后接收错误；nil 时不报告。
	Transceiver    *Transceiver   // 事件收发器。
	RetryTimes     int            // 网络 I/O 超时后的重试次数。
	PayloadHandler PayloadHandler // 业务负载处理器。
}

// SendData 发送一个业务负载事件，并在网络 I/O 超时时重试。
func (t *TransProtocol) SendData(data []byte) error {
	if t.Transceiver == nil {
		return fmt.Errorf("%w: Transceiver is nil", ErrProtocol)
	}

	err := t.retrySend(t.Transceiver.Send(Event[*gtp.MsgPayload]{
		Msg: &gtp.MsgPayload{Data: data},
	}.Interface()))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	return nil
}

func (t *TransProtocol) retrySend(err error) error {
	return Retry{
		Transceiver: t.Transceiver,
		Times:       t.RetryTimes,
	}.Send(err)
}

// HandleEvent 将业务负载事件同步分发给 PayloadHandler，其他消息会被忽略。
func (t *TransProtocol) HandleEvent(e IEvent) {
	switch e.Msg.MsgID() {
	case gtp.MsgID_Payload:
		t.PayloadHandler.Call(t.AutoRecover, t.ReportError, nil, AssertEvent[*gtp.MsgPayload](e))
	}
}
