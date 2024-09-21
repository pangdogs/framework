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
	"github.com/elliotchance/pie/v2"
	"math/rand"
	"slices"
)

func MakeLockedSlice[T any](len, cap int) LockedSlice[T] {
	return LockedSlice[T]{
		RWLocked: MakeRWLocked[[]T](make([]T, len, cap)),
	}
}

func NewLockedSlice[T any](len, cap int) *LockedSlice[T] {
	return &LockedSlice[T]{
		RWLocked: MakeRWLocked[[]T](make([]T, len, cap)),
	}
}

type LockedSlice[T any] struct {
	RWLocked[[]T]
}

func (ls *LockedSlice[T]) Append(values ...T) {
	ls.AutoLock(func(s *[]T) {
		*s = append(*s, values...)
	})
}

func (ls *LockedSlice[T]) Delete(fun generic.Func1[T, bool]) {
	ls.AutoLock(func(s *[]T) {
		*s = slices.DeleteFunc(*s, fun)
	})
}

func (ls *LockedSlice[T]) Any(fun generic.Func1[T, bool]) (ret bool) {
	ls.AutoRLock(func(s *[]T) {
		ret = pie.Any(*s, fun)
	})
	return
}

func (ls *LockedSlice[T]) All(fun generic.Func1[T, bool]) (ret bool) {
	ls.AutoRLock(func(s *[]T) {
		ret = pie.All(*s, fun)
	})
	return
}

func (ls *LockedSlice[T]) Filter(fun generic.Func1[T, bool]) (ret []T) {
	ls.AutoRLock(func(s *[]T) {
		ret = pie.Filter(*s, fun)
	})
	return
}

func (ls *LockedSlice[T]) FilterNot(fun generic.Func1[T, bool]) (ret []T) {
	ls.AutoRLock(func(s *[]T) {
		ret = pie.FilterNot(*s, fun)
	})
	return
}

func (ls *LockedSlice[T]) Sort(fun generic.Func2[T, T, int]) {
	ls.AutoLock(func(s *[]T) {
		slices.SortFunc(*s, fun)
	})
}

func (ls *LockedSlice[T]) SortStable(fun generic.Func2[T, T, int]) {
	ls.AutoLock(func(s *[]T) {
		slices.SortStableFunc(*s, fun)
	})
}

func (ls *LockedSlice[T]) Shuffle() {
	ls.AutoLock(func(s *[]T) {
		rand.Shuffle(len(*s), func(i, j int) {
			(*s)[i], (*s)[j] = (*s)[j], (*s)[i]
		})
	})
}

func (ls *LockedSlice[T]) Len() (l int) {
	ls.AutoRLock(func(s *[]T) {
		l = len(*s)
	})
	return
}

func (ls *LockedSlice[T]) Range(fun generic.Func1[T, bool]) {
	ls.AutoRLock(func(s *[]T) {
		for i := range *s {
			if !fun.Exec((*s)[i]) {
				return
			}
		}
	})
}

func (ls *LockedSlice[T]) Each(fun generic.Action1[T]) {
	ls.AutoRLock(func(s *[]T) {
		for i := range *s {
			fun.Exec((*s)[i])
		}
	})
}
