package framework

import (
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
)

// ReadChan 读取channel
func ReadChan[T any](iec iEC, ch <-chan T) runtime.AsyncRet {
	return core.ReadChan(iec.GetRuntime().Ctx, ch)
}
