package cache_discovery

import (
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/framework/plugins/discovery"
	"time"
)

// Option 所有选项设置器
type Option struct{}

// RegistryOptions 所有选项
type RegistryOptions struct {
	Registry      discovery.IRegistry // 包装的其他registry插件
	RetryInterval time.Duration       // 错误重试的时间间隔
}

// Default 默认值
func (Option) Default() option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		Option{}.Wrap(nil)(options)
		Option{}.RetryInterval(3 * time.Second)(options)
	}
}

// Wrap 包装的其他registry插件
func (Option) Wrap(r discovery.IRegistry) option.Setting[RegistryOptions] {
	return func(o *RegistryOptions) {
		o.Registry = r
	}
}

// RetryInterval 错误重试的时间间隔
func (Option) RetryInterval(t time.Duration) option.Setting[RegistryOptions] {
	return func(o *RegistryOptions) {
		o.RetryInterval = t
	}
}
