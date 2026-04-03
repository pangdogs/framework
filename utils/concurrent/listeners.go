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
	"runtime"
	"slices"
	"sync/atomic"
)

func NewListener[H, M any](handler H, size int) *Listener[H, M] {
	return &Listener[H, M]{
		Handler: handler,
		Inbox:   make(chan M, size),
	}
}

type Listener[H, M any] struct {
	_       noCopy
	Handler H
	Inbox   chan M
}

func NewListeners[H, M any]() *Listeners[H, M] {
	return &Listeners[H, M]{}
}

type Listeners[H, M any] atomic.Pointer[[]*Listener[H, M]]

func (ls *Listeners[H, M]) Add(handler H, size int) *Listener[H, M] {
	pls := (*atomic.Pointer[[]*Listener[H, M]])(ls)
	l := NewListener[H, M](handler, size)
	for {
		var news []*Listener[H, M]
		old := pls.Load()
		if old != nil {
			news = slices.Clone(*old)
		}
		news = append(news, l)
		if pls.CompareAndSwap(old, &news) {
			break
		}
		runtime.Gosched()
	}
	return l
}

func (ls *Listeners[H, M]) Delete(l *Listener[H, M]) {
	pls := (*atomic.Pointer[[]*Listener[H, M]])(ls)
	for {
		var news []*Listener[H, M]
		old := pls.Load()
		if old != nil {
			news = slices.Clone(*old)
		}
		news = slices.DeleteFunc(news, func(exists *Listener[H, M]) bool { return exists == l })
		if pls.CompareAndSwap(old, &news) {
			break
		}
		runtime.Gosched()
	}
}

func (ls *Listeners[H, M]) Load() []*Listener[H, M] {
	pls := (*atomic.Pointer[[]*Listener[H, M]])(ls)
	return *pls.Load()
}

func (ls *Listeners[H, M]) Broadcast(m M) (rejected int) {
	for _, l := range ls.Load() {
		select {
		case l.Inbox <- m:
		default:
			rejected++
		}
	}
	return
}
