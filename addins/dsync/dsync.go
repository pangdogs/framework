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
	"errors"

	"git.golaxy.org/core/utils/option"
)

var (
	// ErrNotAcquired 表示当前句柄没有持有锁，或后端已失去该锁的所有权。
	ErrNotAcquired = errors.New("dsync: lock is not acquired")
	// ErrAlreadyAcquired 表示当前句柄已经在持有或尝试获取锁。
	ErrAlreadyAcquired = errors.New("dsync: lock is already acquired")
)

// IDistSync 创建特定后端的分布式互斥锁。
type IDistSync interface {
	// NewMutex 创建逻辑名称为 name 的分布式锁句柄；创建本身不会获取锁。
	NewMutex(name string, settings ...option.Setting[DistMutexOptions]) IDistMutex
	// Separator 返回该后端组织层级锁名称时使用的分隔符。
	Separator() string
}
