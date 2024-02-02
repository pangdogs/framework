package conf

import (
	"git.golaxy.org/core/util/option"
)

// Option 所有选项设置器
type Option struct{}

// ConfigOptions 所有选项
type ConfigOptions struct {
	AutomaticEnv bool // 是否读取环境变量
	AutoUpdate   bool // 是否热更新
}

// Default 默认值
func (Option) Default() option.Setting[ConfigOptions] {
	return func(options *ConfigOptions) {
		Option{}.AutomaticEnv(false)(options)
		Option{}.AutoUpdate(false)(options)
	}
}

// AutomaticEnv 是否读取环境变量
func (Option) AutomaticEnv(b bool) option.Setting[ConfigOptions] {
	return func(o *ConfigOptions) {
		o.AutomaticEnv = b
	}
}

// AutoUpdate 是否热更新
func (Option) AutoUpdate(b bool) option.Setting[ConfigOptions] {
	return func(o *ConfigOptions) {
		o.AutoUpdate = b
	}
}
