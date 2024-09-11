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

package concurrent

import (
	"context"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/types"
	"time"
)

func newTask[T Resp](fs *Futures, resp T) iTask {
	ctx, cancel := context.WithCancel(context.Background())

	task := &_Task[T]{
		future: Future{
			Finish:  ctx,
			Id:      fs.makeId(),
			futures: fs,
		},
		resp:      resp,
		terminate: cancel,
	}
	fs.tasks.Store(task.future.Id, task)

	return task
}

type iTask interface {
	Future() Future
	Run(ctx context.Context, timeout time.Duration)
	Resolve(ret async.Ret) error
}

type _Task[T Resp] struct {
	future    Future
	resp      T
	terminate context.CancelFunc
}

func (t *_Task[T]) Future() Future {
	return t.future
}

func (t *_Task[T]) Run(ctx context.Context, timeout time.Duration) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-t.future.futures.ctx.Done():
		t.future.futures.Resolve(t.future.Id, async.RetT[any]{Error: ErrFuturesClosed})
	case <-ctx.Done():
		t.future.futures.Resolve(t.future.Id, async.RetT[any]{Error: ErrFutureCanceled})
	case <-timer.C:
		t.future.futures.Resolve(t.future.Id, async.RetT[any]{Error: ErrFutureTimeout})
	case <-t.future.Finish.Done():
		return
	}
}

func (t *_Task[T]) Resolve(ret async.Ret) (retErr error) {
	t.terminate()

	defer func() {
		if err := types.Panic2Err(recover()); err != nil {
			retErr = err
		}
	}()

	return t.resp.Push(ret)
}
