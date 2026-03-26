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
	EventHandler = generic.DelegateVoid1[IEvent] // 消息事件处理器
)

// EventDispatcher 消息事件分发器
type EventDispatcher struct {
	AutoRecover  bool         // panic时是否自动恢复
	ReportError  chan error   // 在开启panic时自动恢复时，将会恢复并将错误写入此error channel
	Transceiver  *Transceiver // 消息事件收发器
	RetryTimes   int          // 网络io超时时的重试次数
	EventHandler EventHandler // 消息事件处理器列表
}

// Dispatch 分发事件
func (d *EventDispatcher) Dispatch(ctx context.Context) error {
	if d.Transceiver == nil {
		return fmt.Errorf("%w: Transceiver is nil", ErrEvent)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	defer d.Transceiver.GC()

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
