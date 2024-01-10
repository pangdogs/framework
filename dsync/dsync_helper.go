package dsync

import (
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util/option"
	"strings"
)

// NewMutex returns a new distributed mutex with given name.
func NewMutex(servCtx service.Context, name string, settings ...option.Setting[DMutexOptions]) IDistMutex {
	return Using(servCtx).NewMutex(name, settings...)
}

// GetSeparator return name path separator.
func GetSeparator(servCtx service.Context) string {
	return Using(servCtx).GetSeparator()
}

// Path return name path.
func Path(servCtx service.Context, elems ...string) string {
	return strings.Join(elems, GetSeparator(servCtx))
}
