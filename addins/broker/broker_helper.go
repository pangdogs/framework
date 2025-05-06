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

package broker

import (
	"context"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/utils/binaryutil"
	"github.com/elliotchance/pie/v2"
)

type (
	ErrorHandler = generic.DelegateVoid1[error] // handling errors
)

// MakeWriteChan creates a new channel for publishing data to a specific topic.
func MakeWriteChan(broker IBroker, topic string, size int, errorHandler ...ErrorHandler) chan<- binaryutil.RecycleBytes {
	if broker == nil {
		exception.Panicf("%w: broker is nil", core.ErrArgs)
	}

	_errorHandler := pie.First(errorHandler)

	ch := make(chan binaryutil.RecycleBytes, size)

	go func() {
		defer func() {
			for bs := range ch {
				bs.Release()
			}
		}()
		for bs := range ch {
			err := broker.Publish(context.Background(), topic, bs.Data())
			bs.Release()
			if err != nil {
				_errorHandler.SafeCall(nil, err)
			}
		}
	}()

	return ch
}

// MakeReadChan creates a new channel for receiving data from a specific pattern.
func MakeReadChan(broker IBroker, ctx context.Context, pattern, queue string, size int) (<-chan binaryutil.RecycleBytes, error) {
	if broker == nil {
		exception.Panicf("%w: broker is nil", core.ErrArgs)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	ch := make(chan binaryutil.RecycleBytes, size)

	_, err := broker.Subscribe(ctx, pattern,
		With.Queue(queue),
		With.EventHandler(generic.CastDelegate1(func(e Event) error {
			bs := binaryutil.MakeNonRecycleBytes(e.Message)

			select {
			case ch <- bs:
				return nil
			default:
				var nakErr error
				if e.Queue != "" {
					nakErr = e.Nak(context.Background())
				}
				return fmt.Errorf("read chan is full, nak: %v", nakErr)
			}
		})),
		With.UnsubscribedCB(func(sub ISubscriber) {
			close(ch)
		}))
	if err != nil {
		return nil, err
	}

	return ch, nil
}
