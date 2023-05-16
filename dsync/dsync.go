package dsync

type DSync interface {
	// NewDMutex returns a new distributed mutex with given name.
	NewDMutex(name string, options ...Option) DMutex
}
