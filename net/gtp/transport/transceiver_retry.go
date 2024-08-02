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
)

// Retry 网络io超时时重试
type Retry struct {
	Transceiver *Transceiver
	Times       int
	Ctx         context.Context
}

// Send 重试发送
func (r Retry) Send(err error) error {
	if err == nil {
		return nil
	}
	if !errors.Is(err, ErrDeadlineExceeded) {
		return err
	}
	ctx := r.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	for i := r.Times; i > 0; i-- {
		select {
		case <-ctx.Done():
			return fmt.Errorf("gtp: %w", context.Canceled)
		default:
		}
		if err = r.Transceiver.Resend(); err != nil {
			if errors.Is(err, ErrDeadlineExceeded) {
				continue
			}
		}
		break
	}
	return err
}

// Recv 重试接收
func (r Retry) Recv(e IEvent, err error) (IEvent, error) {
	if err == nil {
		return e, nil
	}
	if !errors.Is(err, ErrDeadlineExceeded) && !errors.Is(err, ErrDiscardSeq) {
		return e, err
	}
	for i := r.Times; i > 0; {
		e, err = r.Transceiver.Recv(r.Ctx)
		if err != nil {
			if errors.Is(err, ErrDeadlineExceeded) {
				i--
				continue
			}
			if errors.Is(err, ErrDiscardSeq) {
				continue
			}
		}
		break
	}
	return e, err
}
