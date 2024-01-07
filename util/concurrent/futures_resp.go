package concurrent

// Resp 响应接口
type Resp interface {
	// Push 填入返回结果
	Push(ret Ret[any]) error
}

// MakeRet 创建结果
func MakeRet[T any](val T, err error) Ret[T] {
	return Ret[T]{
		Value: val,
		Error: err,
	}
}

// Ret 返回结果
type Ret[T any] struct {
	Value T     // 返回值
	Error error // 返回错误
}

// OK 是否成功
func (ret Ret[T]) OK() bool {
	return ret.Error == nil
}
