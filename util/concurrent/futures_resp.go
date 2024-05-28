package concurrent

import "git.golaxy.org/core/utils/async"

// Resp 响应接口
type Resp interface {
	// Push 填入返回结果
	Push(ret async.Ret) error
}
