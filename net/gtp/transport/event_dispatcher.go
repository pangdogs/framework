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
	"errors"
	"fmt"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/generic"
	"github.com/elliotchance/pie/v2"
)

type (
	EventHandler = generic.DelegateFunc1[IEvent, error] // 消息事件处理器
	ErrorHandler = generic.DelegateAction1[error]       // 错误处理器
)

// EventDispatcher 消息事件分发器
type EventDispatcher struct {
	Transceiver  *Transceiver // 消息事件收发器
	RetryTimes   int          // 网络io超时时的重试次数
	EventHandler EventHandler // 消息事件处理器列表
}

// Dispatching 分发事件
func (d *EventDispatcher) Dispatching(ctx context.Context) error {
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

	var errs []error

	d.EventHandler.Invoke(func(err, panicErr error) bool {
		if err := generic.FuncError(err, panicErr); err != nil {
			errs = append(errs, err)
		}
		return panicErr != nil
	}, e)

	if len(errs) > 0 {
		return fmt.Errorf("%w: %w", ErrEvent, errors.Join(errs...))
	}

	return nil
}

// Run 运行
func (d *EventDispatcher) Run(ctx context.Context, errorHandler ...ErrorHandler) {
	if d.Transceiver == nil {
		exception.Panicf("%w: Transceiver is nil", ErrEvent)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	_errorHandler := pie.First(errorHandler)

	defer d.Transceiver.Clean()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		err := d.Dispatching(ctx)
		if err != nil {
			_errorHandler.Invoke(nil, err)
		}
	}
}

func (d *EventDispatcher) retryRecv(ctx context.Context) (IEvent, error) {
	e, err := d.Transceiver.Recv(ctx)
	return Retry{
		Transceiver: d.Transceiver,
		Times:       d.RetryTimes,
		Ctx:         ctx,
	}.Recv(e, err)
}
