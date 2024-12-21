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

// A IDistMutex is a distributed mutual exclusion lock. Avoid sharing the same IDistMutex instance among multiple goroutines. Create a separate IDistMutex instance for each goroutine.
type IDistMutex interface {
	// Name returns mutex name.
	Name() string
	// Value returns the current random value. The value will be empty until a lock is acquired (or Value option is used).
	Value() string
	// Until returns the time of validity of acquired lock. The value will be zero value until a lock is acquired.
	Until() time.Time
	// Lock locks m. In case it returns an error on failure, you may retry to acquire the lock by calling this method again.
	Lock(ctx context.Context) error
	// Unlock unlocks m and returns the status of unlock.
	Unlock(ctx context.Context) error
	// Extend resets the mutex's expiry and returns the status of expiry extension.
	Extend(ctx context.Context) error
	// Valid returns true if the lock acquired through m is still valid. It may also return true erroneously if quorum is achieved during the call and at
	// least one node then takes long enough to respond for the lock to expire.
	Valid(ctx context.Context) (bool, error)
}
