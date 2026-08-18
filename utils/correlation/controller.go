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

package correlation

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"sync"
	"time"

	"git.golaxy.org/core/utils/async"
)

var (
	// ErrClosed 表示 Controller 已关闭，不再接受新的关联项。
	ErrClosed = errors.New("correlation controller closed")
	// ErrTimeout 表示关联请求在截止时间前没有收到响应。
	ErrTimeout = errors.New("correlation timeout")
)

// ID 是请求与响应之间的不透明关联标识。ID 可随请求跨链路往返，但只对创建它的
// Controller 有意义；零值表示无需关联。
type ID uint64

// New 创建由 ctx 控制生命周期、使用统一超时时长的关联控制器。
// ctx 为 nil 时使用 context.Background。
func New(ctx context.Context, timeout time.Duration) *Controller {
	if ctx == nil {
		ctx = context.Background()
	}

	var seed [16]byte
	rand.Read(seed[:])

	done, _ := async.NewSignal()
	controller := &Controller{
		timeout: timeout,
		pending: make(map[ID]*entry),
		done:    done,
		idSalt:  binary.LittleEndian.Uint64(seed[:8]),
		idMask:  binary.LittleEndian.Uint64(seed[8:]),
	}
	stopWatch := context.AfterFunc(ctx, func() {
		controller.Close()
	})

	controller.mu.Lock()
	if controller.closed {
		controller.mu.Unlock()
		stopWatch()
	} else {
		controller.stopWatch = stopWatch
		controller.mu.Unlock()
	}
	if ctx.Err() != nil {
		controller.Close()
	}
	return controller
}

// Controller 管理异步请求的登记、响应匹配、取消、超时和整体关闭。
// Controller 初始化后不得复制。
type Controller struct {
	_         noCopy
	mu        sync.Mutex
	closed    bool
	timeout   time.Duration
	pending   map[ID]*entry
	done      async.Completer
	stopWatch func() bool
	idSalt    uint64
	idMask    uint64
}

// Begin 登记新的待响应请求，并返回关联 ID 和只读 Future。关联 ID 由 Future ID
// 经过 Controller 私有参数混淆得到；Controller 只在内部持有对应 Promise。
// Controller 已关闭时返回 ErrClosed。
func (controller *Controller) Begin() (ID, async.Future, error) {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	if controller.closed {
		return 0, async.Future{}, ErrClosed
	}

	var promise async.Promise
	var future async.Future
	var id ID
	for {
		promise, future = async.NewPromise()
		id = controller.makeID(future.ID())
		if id != 0 {
			if _, exists := controller.pending[id]; !exists {
				break
			}
		}
	}
	item := &entry{
		promise:  promise,
		deadline: time.Now().Add(controller.timeout),
	}
	controller.pending[id] = item
	item.timer = time.AfterFunc(controller.timeout, func() {
		controller.complete(id, item, async.NewResult(nil, ErrTimeout), false)
	})

	return id, future, nil
}

// Resolve 以 ret 完成关联请求。请求不存在、已经完成或已经超时时返回 false。
func (controller *Controller) Resolve(id ID, ret async.Result) bool {
	return controller.complete(id, nil, ret, true)
}

// Cancel 以 cause 取消关联请求。请求不存在、已经完成或已经超时时返回 false。
func (controller *Controller) Cancel(id ID, cause error) bool {
	if cause == nil {
		cause = context.Canceled
	}
	return controller.complete(id, nil, async.NewResult(nil, cause), true)
}

// Close 幂等关闭 Controller，并以 ErrClosed 完成所有尚未结束的请求。
// 首次关闭返回 true；Controller 已关闭时返回 false。
func (controller *Controller) Close() bool {
	controller.mu.Lock()
	if controller.closed {
		controller.mu.Unlock()
		return false
	}
	controller.closed = true
	pending := controller.pending
	controller.pending = nil
	stopWatch := controller.stopWatch
	controller.stopWatch = nil
	controller.mu.Unlock()

	if stopWatch != nil {
		stopWatch()
	}
	for _, item := range pending {
		item.timer.Stop()
		item.promise.Resolve(async.NewResult(nil, ErrClosed))
	}
	controller.done.Complete()
	return true
}

// Done 返回 Controller 完成关闭和待处理请求收尾时兑现的 Signal。
func (controller *Controller) Done() async.Signal {
	return controller.done.Signal()
}

// makeID 以 64 位双射混淆 Future ID；不同的非环回输入不会产生碰撞。
func (controller *Controller) makeID(futureID async.FutureID) ID {
	z := uint64(futureID) + controller.idSalt
	z = (z ^ z>>30) * 0xbf58476d1ce4e5b9
	z = (z ^ z>>27) * 0x94d049bb133111eb
	return ID((z ^ z>>31) ^ controller.idMask)
}

func (controller *Controller) complete(id ID, expected *entry, ret async.Result, checkDeadline bool) bool {
	controller.mu.Lock()
	item := controller.pending[id]
	if item == nil || expected != nil && item != expected {
		controller.mu.Unlock()
		return false
	}

	completedInTime := true
	if checkDeadline && !time.Now().Before(item.deadline) {
		ret = async.NewResult(nil, ErrTimeout)
		completedInTime = false
	}
	controller.mu.Unlock()

	item.timer.Stop()
	if !item.promise.Resolve(ret) {
		return false
	}

	controller.mu.Lock()
	if controller.pending[id] == item {
		delete(controller.pending, id)
	}
	controller.mu.Unlock()
	return completedInTime
}

type entry struct {
	promise  async.Promise
	deadline time.Time
	timer    *time.Timer
}
