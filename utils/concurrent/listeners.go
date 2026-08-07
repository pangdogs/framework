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

// NewListener 创建携带 handler、Inbox 容量为 size 的监听者；size 为负数时 panic。
func NewListener[H, M any](handler H, size int) *Listener[H, M] {
	return &Listener[H, M]{
		Handler: handler,
		Inbox:   make(chan M, size),
	}
}

// Listener 将调用方定义的 handler 与接收广播消息的 Inbox 关联起来。
// Listener 初始化后不得复制；Inbox 的消费由调用方负责。
type Listener[H, M any] struct {
	_       noCopy
	Handler H      // 与监听者关联的处理器或元数据。
	Inbox   chan M // 接收广播消息的有缓冲通道。
}

// NewListeners 创建空监听者集合；Listeners 的零值同样可直接使用。
func NewListeners[H, M any]() *Listeners[H, M] {
	return &Listeners[H, M]{}
}

// Listeners 以原子写时复制快照维护监听者集合。
// 添加、删除、加载和广播可并发调用；集合开始使用后不得复制。
type Listeners[H, M any] atomic.Pointer[[]*Listener[H, M]]

// Add 创建并加入监听者，返回值用于接收消息和后续 Delete；相同 handler 可重复添加。
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

// Delete 从后续快照中删除指定监听者，但不关闭其 Inbox。
// 已取得旧快照的并发广播仍可能向该监听者投递一次消息。
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

// Load 返回当前监听者快照；调用方必须将返回切片视为只读。
func (ls *Listeners[H, M]) Load() []*Listener[H, M] {
	pls := (*atomic.Pointer[[]*Listener[H, M]])(ls)
	snap := pls.Load()
	if snap == nil {
		return nil
	}
	return *snap
}

// Broadcast 尝试以非阻塞方式向当前快照中的每个 Inbox 投递 m，并返回因通道已满而拒绝的数量。
// 调用方不得关闭仍在集合或并发广播旧快照中的 Inbox，否则发送会 panic。
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
