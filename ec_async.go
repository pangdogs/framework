package framework

import (
	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/async"
)

// ReadChan 读取channel
func ReadChan[T any](iec iEC, ch <-chan T) async.AsyncRet {
	return core.ReadChan(iec.GetRuntime(), ch)
}
