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
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
)

var (
	ErrFutureControllerClosed = errors.New("future controller closed")
	ErrFutureExceeded         = errors.New("future exceeded deadline")
)

func NewFutureController(ctx context.Context, timeout time.Duration) *FutureController {
	if ctx == nil {
		ctx = context.Background()
	}

	fc := &FutureController{
		ctx:            ctx,
		terminated:     async.NewFutureVoid(),
		timeout:        timeout,
		pendingResolve: make(map[int64]*FutureHandle),
		pendingTimeout: generic.NewUnboundedChannel[*FutureHandle](),
	}
	fc.idGen.Store(rand.Int63())

	go fc.watchingForTimeout()

	return fc
}

type FutureController struct {
	_              noCopy
	ctx            context.Context
	terminated     async.FutureVoid
	barrier        generic.Barrier
	idGen          atomic.Int64
	timeout        time.Duration
	mutex          sync.Mutex
	pendingResolve map[int64]*FutureHandle
	pendingTimeout *generic.UnboundedChannel[*FutureHandle]
}

func (fc *FutureController) New() (*FutureHandle, error) {
	select {
	case <-fc.ctx.Done():
		return nil, ErrFutureControllerClosed
	default:
	}

	if !fc.barrier.Join(1) {
		return nil, ErrFutureControllerClosed
	}

	handle := &FutureHandle{
		id:         fc.genId(),
		future:     async.NewFutureChan(),
		deadline:   time.Now().Add(fc.timeout),
		controller: fc,
	}

	fc.mutex.Lock()
	fc.pendingResolve[handle.id] = handle
	fc.mutex.Unlock()

	fc.pendingTimeout.In() <- handle

	fc.barrier.Done()

	return handle, nil
}

func (fc *FutureController) Resolve(id int64, ret async.Result) error {
	fc.mutex.Lock()
	handle := fc.pendingResolve[id]
	if handle == nil {
		fc.mutex.Unlock()
		return ErrFutureExceeded
	}
	if !time.Now().Before(handle.deadline) {
		fc.mutex.Unlock()
		return ErrFutureExceeded
	}
	delete(fc.pendingResolve, id)
	fc.mutex.Unlock()

	if !handle.Resolved.CompareAndSwap(false, true) {
		return ErrFutureExceeded
	}

	async.Return(handle.future, ret)
	return nil
}

func (fc *FutureController) Terminated() async.Future {
	return fc.terminated.Out()
}

func (fc *FutureController) watchingForTimeout() {
loop:
	for {
		select {
		case <-fc.ctx.Done():
			break loop
		case handle := <-fc.pendingTimeout.Out():
			if delta := time.Until(handle.deadline); delta > 0 {
				time.Sleep(delta)
			}

			if !handle.Resolved.CompareAndSwap(false, true) {
				continue
			}

			fc.mutex.Lock()
			delete(fc.pendingResolve, handle.id)
			fc.mutex.Unlock()

			async.Return(handle.future, async.NewResult(nil, ErrFutureExceeded))
		}
	}

	fc.barrier.Close()
	fc.barrier.Wait()

	fc.pendingTimeout.Close()

	for handle := range fc.pendingTimeout.Out() {
		fc.mutex.Lock()
		delete(fc.pendingResolve, handle.id)
		fc.mutex.Unlock()

		if !handle.Resolved.CompareAndSwap(false, true) {
			continue
		}

		async.Return(handle.future, async.NewResult(nil, ErrFutureControllerClosed))
	}

	async.ReturnVoid(fc.terminated)
}

func (fc *FutureController) genId() int64 {
	id := fc.idGen.Add(1)
	if id == 0 {
		id = fc.idGen.Add(1)
	}
	return id
}
