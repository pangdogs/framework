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
	"git.golaxy.org/core/utils/generic"
	"sync"
)

func MakeRWLocked[T any](obj T) RWLocked[T] {
	return RWLocked[T]{
		object: obj,
	}
}

func NewRWLocked[T any](obj T) *RWLocked[T] {
	return &RWLocked[T]{
		object: obj,
	}
}

type RWLocked[T any] struct {
	object T
	mutex  sync.RWMutex
}

func (l *RWLocked[T]) AutoLock(fun generic.Action1[*T]) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	fun.Exec(&l.object)
}

func (l *RWLocked[T]) AutoRLock(fun generic.Action1[*T]) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	fun.Exec(&l.object)
}
