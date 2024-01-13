package dsync

import (
	"errors"
	"fmt"
	"git.golaxy.org/core/util/option"
)

var (
	// ErrDsync dsync errors.
	ErrDsync = errors.New("dsync")
	// ErrNotAcquired is an error indicating that the distributed lock was not acquired. It is returned by IDistMutex.Unlock and IDistMutex.Extend when the lock was not successfully acquired or has expired.
	ErrNotAcquired = fmt.Errorf("%w: lock is not acquired", ErrDsync)
)

// IDistSync represents a distributed synchronization mechanism.
type IDistSync interface {
	// NewMutex returns a new distributed mutex with given name.
	NewMutex(name string, settings ...option.Setting[DMutexOptions]) IDistMutex
	// GetSeparator return name path separator.
	GetSeparator() string
}
