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

// Retry 为 Transceiver 的超时或重复序号错误提供有限次重试。
type Retry struct {
	Transceiver *Transceiver    // 被重试的收发器；发生重试时不得为 nil。
	Times       int             // 最大重试次数；小于等于零时直接返回原错误。
	Ctx         context.Context // 控制重试过程；nil 表示 context.Background()。
}

// Send 在 err 表示 I/O 超时时调用 Resend，最多重试 Times 次；其他错误原样返回。
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

// Recv 在 err 表示 I/O 超时或重复序号时继续接收。
// 超时会消耗一次重试次数，重复序号只会丢弃当前包，不消耗重试次数。
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
