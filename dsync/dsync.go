package dsync

import (
	"errors"
	"fmt"
	"kit.golaxy.org/golaxy/util/option"
)

var (
	// ErrDsync dsync errors.
	ErrDsync = errors.New("dsync")
	// ErrNotAcquired is an error indicating that the distributed lock was not acquired. It is returned by DMutex.Unlock and DMutex.Extend when the lock was not successfully acquired or has expired.
	ErrNotAcquired = fmt.Errorf("%w: lock is not acquired", ErrDsync)
)

// DSync represents a distributed synchronization mechanism.
type DSync interface {
	// NewMutex returns a new distributed mutex with given name.
	NewMutex(name string, settings ...option.Setting[DMutexOptions]) DMutex
	// Separator return name path separator.
	Separator() string
}
