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
	// ErrNotAcquired is an error indicating that the distributed lock was not acquired. It is returned by IDistMutex.Unlock and IDistMutex.Extend when the lock was not successfully acquired or has expired.
	ErrNotAcquired = errors.New("dsync: lock is not acquired")
)

// IDistMutexSettings represents an interface for configuring a distributed mutex.
type IDistMutexSettings interface {
	// With applies additional settings to the distributed mutex.
	With(settings ...option.Setting[DistMutexOptions]) IDistMutex
}

// IDistSync represents a distributed synchronization mechanism.
type IDistSync interface {
	// NewMutex returns a new distributed mutex with given name.
	NewMutex(name string, settings ...option.Setting[DistMutexOptions]) IDistMutex
	// NewMutexf returns a new distributed mutex using a formatted string.
	NewMutexf(format string, args ...any) IDistMutexSettings
	// NewMutexp returns a new distributed mutex using elements.
	NewMutexp(elems ...string) IDistMutexSettings
	// GetSeparator return name path separator.
	GetSeparator() string
}
