package dsync

import "kit.golaxy.org/golaxy/service"

// NewMutex returns a new distributed mutex with given name.
func NewMutex(serviceCtx service.Context, name string, options ...DMutexOption) DMutex {
	return Fetch(serviceCtx).NewMutex(name, options...)
}

// Separator return topic path separator.
func Separator(serviceCtx service.Context) string {
	return Fetch(serviceCtx).Separator()
}
