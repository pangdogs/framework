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

import "git.golaxy.org/core/utils/generic"

func MakeLockedMap[K comparable, V any](size int) LockedMap[K, V] {
	return LockedMap[K, V]{
		RWLocked: MakeRWLocked(make(map[K]V, size)),
	}
}

func NewLockedMap[K comparable, V any](size int) *LockedMap[K, V] {
	return &LockedMap[K, V]{
		RWLocked: MakeRWLocked(make(map[K]V, size)),
	}
}

type LockedMap[K comparable, V any] struct {
	RWLocked[map[K]V]
}

func (lm *LockedMap[K, V]) Add(k K, v V) {
	lm.AutoLock(func(m *map[K]V) {
		(*m)[k] = v
	})
}

func (lm *LockedMap[K, V]) TryAdd(k K, v V) {
	lm.AutoLock(func(m *map[K]V) {
		if _, ok := (*m)[k]; ok {
			return
		}
		(*m)[k] = v
	})
}

func (lm *LockedMap[K, V]) Delete(k K) {
	lm.AutoLock(func(m *map[K]V) {
		delete(*m, k)
	})
}

func (lm *LockedMap[K, V]) Get(k K) (v V, ok bool) {
	lm.AutoRLock(func(m *map[K]V) {
		v, ok = (*m)[k]
	})
	return
}

func (lm *LockedMap[K, V]) Value(k K) (v V) {
	lm.AutoRLock(func(m *map[K]V) {
		v, _ = (*m)[k]
	})
	return
}

func (lm *LockedMap[K, V]) Exist(k K) (b bool) {
	lm.AutoRLock(func(m *map[K]V) {
		_, b = (*m)[k]
	})
	return
}

func (lm *LockedMap[K, V]) Len() (l int) {
	lm.AutoRLock(func(m *map[K]V) {
		l = len(*m)
	})
	return
}

func (lm *LockedMap[K, V]) Range(fun generic.Func2[K, V, bool]) {
	lm.AutoRLock(func(m *map[K]V) {
		for k, v := range *m {
			if !fun.UnsafeCall(k, v) {
				return
			}
		}
	})
}

func (lm *LockedMap[K, V]) Each(fun generic.Action2[K, V]) {
	lm.AutoRLock(func(m *map[K]V) {
		for k, v := range *m {
			fun.UnsafeCall(k, v)
		}
	})
}
