package registry

import "time"

// RegisterOptions 注册函数的所有选项
type RegisterOptions struct {
	TTL time.Duration // 超时时间
}

// RegisterOption 注册函数的选项设置器
type RegisterOption func(o *RegisterOptions)

// WithRegisterOption 注册函数的所有选项设置器
type WithRegisterOption struct{}

// Default 默认值
func (WithRegisterOption) Default() RegisterOption {
	return func(o *RegisterOptions) {
		WithRegisterOption{}.TTL(3 * time.Second)(o)
	}
}

// TTL 超时时间
func (WithRegisterOption) TTL(ttl time.Duration) RegisterOption {
	return func(o *RegisterOptions) {
		o.TTL = ttl
	}
}
