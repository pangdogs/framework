package dsync

import (
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util/option"
	"strings"
)

// NewMutex returns a new distributed mutex with given name.
func NewMutex(servCtx service.Context, name string, settings ...option.Setting[DMutexOptions]) DMutex {
	return Using(servCtx).NewMutex(name, settings...)
}

// Separator return name path separator.
func Separator(servCtx service.Context) string {
	return Using(servCtx).Separator()
}

// Path return name path.
func Path(servCtx service.Context, elems ...string) string {
	return strings.Join(elems, Separator(servCtx))
}
