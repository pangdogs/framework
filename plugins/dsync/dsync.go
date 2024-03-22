package dsync

import (
	"errors"
	"git.golaxy.org/core/util/option"
)

var (
	// ErrNotAcquired is an error indicating that the distributed lock was not acquired. It is returned by IDistMutex.Unlock and IDistMutex.Extend when the lock was not successfully acquired or has expired.
	ErrNotAcquired = errors.New("dsync: lock is not acquired")
)

// IDistSync represents a distributed synchronization mechanism.
type IDistSync interface {
	// NewMutex returns a new distributed mutex with given name.
	NewMutex(name string, settings ...option.Setting[DistMutexOptions]) IDistMutex
	// GetSeparator return name path separator.
	GetSeparator() string
}
