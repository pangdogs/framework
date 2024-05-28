package cache_discovery

import (
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/plugins/discovery"
)

// RegistryOptions 所有选项
type RegistryOptions struct {
	Registry discovery.IRegistry // 包装的其他registry插件
}

var With _Option

type _Option struct{}

// Default 默认值
func (_Option) Default() option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		With.Wrap(nil)(options)
	}
}

// Wrap 包装的其他registry插件
func (_Option) Wrap(r discovery.IRegistry) option.Setting[RegistryOptions] {
	return func(o *RegistryOptions) {
		o.Registry = r
	}
}
