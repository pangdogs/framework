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

package dsync

import (
	"context"
	"time"
)

// IDistMutex 表示一个具有实现特定租约语义的分布式互斥锁。
// 同一个句柄不能并发执行加锁、解锁或续期操作。
type IDistMutex interface {
	// Name 返回不含后端键前缀的锁名称。
	Name() string
	// UID 返回当前锁所有权标识；后端不支持或尚未加锁时可能为空。
	UID() string
	// Until 返回当前租约的预计失效时间；后端不支持时返回零值。
	Until() time.Time
	// TryLock 尝试获取锁；重试策略由具体后端和 Option 决定。
	TryLock(ctx context.Context) error
	// Lock 等待并获取锁；等待受 ctx 及具体后端超时策略约束。
	Lock(ctx context.Context) error
	// Unlock 释放当前持有的锁。
	Unlock(ctx context.Context) error
	// Extend 延长当前锁租约；不支持续期的后端会返回错误。
	Extend(ctx context.Context) error
}
