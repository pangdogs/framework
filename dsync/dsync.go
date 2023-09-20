package dsync

// DSync represents a distributed synchronization mechanism.
type DSync interface {
	// NewMutex returns a new distributed mutex with given name.
	NewMutex(name string, options ...DMutexOption) DMutex
	// Separator return name path separator.
	Separator() string
}
