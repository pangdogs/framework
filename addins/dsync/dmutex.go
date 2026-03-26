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

// IDistMutex 分布式锁接口
type IDistMutex interface {
	// Name 名称
	Name() string
	// UID 唯一ID
	UID() string
	// Until 返回锁的有效期结束时间
	Until() time.Time
	// TryLock 尝试加锁，支持错误重试
	TryLock(ctx context.Context) error
	// Lock 加锁，支持错误重试
	Lock(ctx context.Context) error
	// Unlock 解锁
	Unlock(ctx context.Context) error
	// Extend 延长锁的过期时间
	Extend(ctx context.Context) error
}
