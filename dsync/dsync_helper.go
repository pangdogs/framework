package dsync

import "kit.golaxy.org/golaxy/service"

// NewDMutex returns a new distributed mutex with given name.
func NewDMutex(serviceCtx service.Context, name string, options ...DMutexOption) DMutex {
	return Fetch(serviceCtx).NewDMutex(name, options...)
}
