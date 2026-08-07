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
	// ErrFutureControllerClosed 表示控制器的父 context 已结束，不再接受或等待任务。
	ErrFutureControllerClosed = errors.New("future controller closed")
	// ErrFutureExceeded 表示 Future 已超过截止时间、已完成或不存在。
	ErrFutureExceeded = errors.New("future exceeded deadline")
)

// NewFutureController 创建以 ctx 控制生命周期、以 timeout 设置每个 Future 截止时间的控制器。
// ctx 为 nil 时使用 context.Background；调用方应取消 ctx 以终止内部 watcher。
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

// FutureController 并发管理带唯一 ID 和统一超时时长的 Future，并将 Resolve 与等待方匹配。
// 控制器初始化后不得复制。
type FutureController struct {
	_                noCopy
	ctx              context.Context
	terminated       async.FutureVoid
	barrier          generic.Barrier
	idGen            atomic.Int64
	timeout          time.Duration
	pendingResolveMu sync.Mutex
	pendingResolve   map[int64]*FutureHandle
	pendingTimeout   *generic.UnboundedChannel[*FutureHandle]
}

// New 注册一个待完成的 Future；控制器已结束时返回 ErrFutureControllerClosed。
func (fc *FutureController) New() (*FutureHandle, error) {
	select {
	case <-fc.ctx.Done():
		return nil, ErrFutureControllerClosed
	default:
	}

	if !fc.barrier.Join(1) {
		return nil, ErrFutureControllerClosed
	}
	defer fc.barrier.Done()

	handle := &FutureHandle{
		id:         fc.genId(),
		future:     async.NewFutureChan(),
		deadline:   time.Now().Add(fc.timeout),
		controller: fc,
	}

	fc.pendingResolveMu.Lock()
	fc.pendingResolve[handle.id] = handle
	fc.pendingResolveMu.Unlock()

	fc.pendingTimeout.In() <- handle

	return handle, nil
}

// Resolve 以 ret 完成指定 Future。
// ID 不存在、Future 已完成或已到截止时间时返回 ErrFutureExceeded。
func (fc *FutureController) Resolve(id int64, ret async.Result) error {
	fc.pendingResolveMu.Lock()
	handle := fc.pendingResolve[id]
	if handle == nil {
		fc.pendingResolveMu.Unlock()
		return ErrFutureExceeded
	}
	if !time.Now().Before(handle.deadline) {
		fc.pendingResolveMu.Unlock()
		return ErrFutureExceeded
	}
	delete(fc.pendingResolve, id)
	fc.pendingResolveMu.Unlock()

	if !handle.resolved.CompareAndSwap(false, true) {
		return ErrFutureExceeded
	}

	async.Return(handle.future, ret)
	return nil
}

// Terminated 返回控制器结束 Future；父 context 取消且所有待处理项完成收尾后该 Future 才会完成。
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

			if !handle.resolved.CompareAndSwap(false, true) {
				continue
			}

			fc.pendingResolveMu.Lock()
			delete(fc.pendingResolve, handle.id)
			fc.pendingResolveMu.Unlock()

			async.Return(handle.future, async.NewResult(nil, ErrFutureExceeded))
		}
	}

	fc.barrier.Close()
	fc.barrier.Wait()

	fc.pendingTimeout.Close()

	for handle := range fc.pendingTimeout.Out() {
		fc.pendingResolveMu.Lock()
		delete(fc.pendingResolve, handle.id)
		fc.pendingResolveMu.Unlock()

		if !handle.resolved.CompareAndSwap(false, true) {
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
