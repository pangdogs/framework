package dsync

import (
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util/option"
)

// NewMutex returns a new distributed mutex with given name.
func NewMutex(serviceCtx service.Context, name string, settings ...option.Setting[DMutexOptions]) DMutex {
	return Using(serviceCtx).NewMutex(name, settings...)
}

// Separator return topic path separator.
func Separator(serviceCtx service.Context) string {
	return Using(serviceCtx).Separator()
}
