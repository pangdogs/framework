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

type FutureHandle struct {
	_          noCopy
	id         int64
	future     async.FutureChan
	deadline   time.Time
	resolved   atomic.Bool
	controller *FutureController
}

func (h *FutureHandle) Id() int64 {
	return h.id
}

func (h *FutureHandle) Future() async.Future {
	return h.future.Out()
}

func (h *FutureHandle) Deadline() time.Time {
	return h.deadline
}

func (h *FutureHandle) Cancel(err error) {
	h.controller.Resolve(h.id, async.NewResult(nil, err))
}

func (h *FutureHandle) Resolve(ret async.Result) error {
	return h.controller.Resolve(h.id, ret)
}
