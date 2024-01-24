package cache_discovery

import (
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/framework/plugins/discovery"
)

// Option 所有选项设置器
type Option struct{}

// RegistryOptions 所有选项
type RegistryOptions struct {
	Registry discovery.IRegistry
}

// Default 默认值
func (Option) Default() option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		Option{}.Wrap(nil)(options)
	}
}

// Wrap 包装其他registry插件
func (Option) Wrap(r discovery.IRegistry) option.Setting[RegistryOptions] {
	return func(o *RegistryOptions) {
		o.Registry = r
	}
}
