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

package fanout

import (
	"runtime"
	"slices"
	"sync/atomic"
)

// NewBroadcaster 创建空广播器；Broadcaster 的零值同样可直接使用。
func NewBroadcaster[H, M any]() *Broadcaster[H, M] {
	return &Broadcaster[H, M]{}
}

// Subscription 表示一次广播订阅，将调用方定义的 Handler 与独立 Inbox 关联起来。
// Subscription 初始化后不得复制；Inbox 的消费由调用方负责。
type Subscription[H, M any] struct {
	_       noCopy
	Handler H      // 与订阅关联的处理器或元数据。
	Inbox   chan M // 接收广播消息的有缓冲通道。
}

// Broadcaster 以原子写时复制快照维护订阅集合，并向每个订阅独立投递消息。
// 订阅、退订、快照读取和广播可并发调用；Broadcaster 开始使用后不得复制。
type Broadcaster[H, M any] atomic.Pointer[[]*Subscription[H, M]]

// Subscribe 创建具有指定 Inbox 容量的订阅并加入广播器；capacity 为负数时 panic。
// 相同 Handler 可以重复订阅。
func (b *Broadcaster[H, M]) Subscribe(handler H, capacity int) *Subscription[H, M] {
	pointer := (*atomic.Pointer[[]*Subscription[H, M]])(b)
	sub := &Subscription[H, M]{
		Handler: handler,
		Inbox:   make(chan M, capacity),
	}

	for {
		var next []*Subscription[H, M]
		current := pointer.Load()
		if current != nil {
			next = slices.Clone(*current)
		}
		next = append(next, sub)
		if pointer.CompareAndSwap(current, &next) {
			return sub
		}
		runtime.Gosched()
	}
}

// Unsubscribe 从后续快照中删除指定订阅，但不关闭其 Inbox。
// 已取得旧快照的并发广播仍可能向该订阅投递一次消息。
func (b *Broadcaster[H, M]) Unsubscribe(sub *Subscription[H, M]) {
	pointer := (*atomic.Pointer[[]*Subscription[H, M]])(b)

	for {
		var next []*Subscription[H, M]
		current := pointer.Load()
		if current != nil {
			next = slices.Clone(*current)
		}
		next = slices.DeleteFunc(next, func(item *Subscription[H, M]) bool { return item == sub })
		if pointer.CompareAndSwap(current, &next) {
			return
		}
		runtime.Gosched()
	}
}

// Snapshot 返回当前订阅快照；调用方必须将返回切片视为只读。
func (b *Broadcaster[H, M]) Snapshot() []*Subscription[H, M] {
	pointer := (*atomic.Pointer[[]*Subscription[H, M]])(b)
	snapshot := pointer.Load()
	if snapshot == nil {
		return nil
	}
	return *snapshot
}

// Broadcast 尝试以非阻塞方式向当前快照中的每个 Inbox 投递 message，并返回因 Inbox
// 已满而丢弃的投递数量。调用方不得关闭仍在广播器或并发广播旧快照中的 Inbox。
func (b *Broadcaster[H, M]) Broadcast(message M) (dropped int) {
	for _, sub := range b.Snapshot() {
		select {
		case sub.Inbox <- message:
		default:
			dropped++
		}
	}
	return
}
