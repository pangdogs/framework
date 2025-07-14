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
	"errors"
	"fmt"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/net/gtp"
)

type (
	PayloadHandler = generic.Delegate1[Event[*gtp.MsgPayload], error] // Payload消息事件处理器
)

// TransProtocol 传输协议
type TransProtocol struct {
	Transceiver    *Transceiver   // 消息事件收发器
	RetryTimes     int            // 网络io超时时的重试次数
	PayloadHandler PayloadHandler // Payload消息事件处理器
}

// SendData 发送数据
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

// HandleRecvEvent 消息事件处理器
func (t *TransProtocol) HandleRecvEvent(e IEvent) error {
	switch e.Msg.MsgId() {
	case gtp.MsgId_Payload:
		var errs []error

		t.PayloadHandler.UnsafeCall(func(err, _ error) bool {
			if err != nil {
				errs = append(errs, err)
			}
			return false
		}, AssertEvent[*gtp.MsgPayload](e))

		if len(errs) > 0 {
			return errors.Join(errs...)
		}

		return nil

	default:
		return nil
	}
}
