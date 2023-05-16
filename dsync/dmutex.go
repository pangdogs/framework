package dsync

import (
	"context"
	"time"
)

// A DMutex is a distributed mutual exclusion lock.
type DMutex interface {
	// Name returns mutex name (i.e. the Redis key).
	Name() string

	// Value returns the current random value. The value will be empty until a lock is acquired (or Value option is used).
	Value() string

	// Until returns the time of validity of acquired lock. The value will be zero value until a lock is acquired.
	Until() time.Time

	// Lock locks m. In case it returns an error on failure, you may retry to acquire the lock by calling this method again.
	Lock(ctx context.Context) error

	// Unlock unlocks m and returns the status of unlock.
	Unlock(ctx context.Context) (bool, error)

	// Extend resets the mutex's expiry and returns the status of expiry extension.
	Extend(ctx context.Context) (bool, error)

	// Valid returns true if the lock acquired through m is still valid. It may
	// also return true erroneously if quorum is achieved during the call and at
	// least one node then takes long enough to respond for the lock to expire.
	Valid(ctx context.Context) (bool, error)
}
