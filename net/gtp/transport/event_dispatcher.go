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

	"git.golaxy.org/core/utils/generic"
)

type (
	// EventHandler 处理一个类型擦除的 GTP 事件。
	EventHandler = generic.DelegateVoid1[IEvent]
)

// EventDispatcher 每次接收并分发一个 GTP 事件。
type EventDispatcher struct {
	AutoRecover  bool         // 是否恢复事件处理器的 panic。
	ReportError  chan error   // 恢复 panic 后接收错误；nil 时不报告。
	Transceiver  *Transceiver // 事件收发器。
	RetryTimes   int          // 网络 I/O 超时后的重试次数。
	EventHandler EventHandler // 事件处理器。
}

// Dispatch 接收一个事件并同步调用处理器；ctx 为 nil 时使用 Background。
func (d *EventDispatcher) Dispatch(ctx context.Context) error {
	if d.Transceiver == nil {
		return fmt.Errorf("%w: Transceiver is nil", ErrEvent)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	e, err := d.retryRecv(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrEvent, err)
	}

	d.EventHandler.Call(d.AutoRecover, d.ReportError, nil, e)

	return nil
}

func (d *EventDispatcher) retryRecv(ctx context.Context) (IEvent, error) {
	e, err := d.Transceiver.Recv(ctx)
	return Retry{
		Transceiver: d.Transceiver,
		Times:       d.RetryTimes,
		Ctx:         ctx,
	}.Recv(e, err)
}
