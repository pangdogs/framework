package cache_registry

import (
	"kit.golaxy.org/plugins/registry"
)

// Option 所有选项设置器
type Option struct{}

// RegistryOptions 所有选项
type RegistryOptions struct {
	Registry registry.Registry
}

// RegistryOption 选项设置器
type RegistryOption func(options *RegistryOptions)

// Default 默认值
func (Option) Default() RegistryOption {
	return func(options *RegistryOptions) {
		Option{}.Wrap(nil)(options)
	}
}

// Wrap 包装其他registry插件
func (Option) Wrap(r registry.Registry) RegistryOption {
	return func(o *RegistryOptions) {
		o.Registry = r
	}
}
