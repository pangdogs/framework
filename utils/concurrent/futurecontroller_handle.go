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
	"sync/atomic"
	"time"

	"git.golaxy.org/core/utils/async"
)

// FutureHandle 标识由 FutureController 管理的一次异步结果等待。
// Handle 初始化后不得复制，完成、取消与超时之间只会有一个结果生效。
type FutureHandle struct {
	_          noCopy
	id         int64
	future     async.FutureChan
	deadline   time.Time
	resolved   atomic.Bool
	controller *FutureController
}

// Id 返回用于匹配响应的唯一标识；零值不会被分配。
func (h *FutureHandle) Id() int64 {
	return h.id
}

// Future 返回用于等待结果的只读 Future。
func (h *FutureHandle) Future() async.Future {
	return h.future.Out()
}

// Deadline 返回该 Future 的截止时间。
func (h *FutureHandle) Deadline() time.Time {
	return h.deadline
}

// Cancel 尝试以 err 完成 Future；Future 已结束时调用无效果，失败不会返回给调用方。
func (h *FutureHandle) Cancel(err error) {
	h.controller.Resolve(h.id, async.NewResult(nil, err))
}

// Resolve 以 ret 完成 Future；已完成或超时时返回 ErrFutureExceeded。
func (h *FutureHandle) Resolve(ret async.Result) error {
	return h.controller.Resolve(h.id, ret)
}
