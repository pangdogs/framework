package conf

import (
	"git.golaxy.org/core/util/option"
)

// Option 所有选项设置器
type Option struct{}

// ConfigOptions 所有选项
type ConfigOptions struct {
	Format         string         // 配置格式（json,yaml,ini...）
	LocalPath      string         // 本地配置文件路径
	RemoteProvider string         // 远端配置类型（etcd3,consul...）
	RemoteEndpoint string         // 远端地址
	RemotePath     string         // 远端路径
	AutoUpdate     bool           // 是否热更新
	DefaultKVs     map[string]any // 默认配置
}

// Default 默认值
func (Option) Default() option.Setting[ConfigOptions] {
	return func(options *ConfigOptions) {
		Option{}.Format("json")(options)
		Option{}.LocalPath("")(options)
		Option{}.Remote("", "", "")(options)
		Option{}.AutoUpdate(false)(options)
		Option{}.DefaultKVs(nil)(options)
	}
}

// Format 配置格式（json,yaml,ini...）
func (Option) Format(format string) option.Setting[ConfigOptions] {
	return func(o *ConfigOptions) {
		o.Format = format
	}
}

// LocalPath 本地配置文件路径
func (Option) LocalPath(path string) option.Setting[ConfigOptions] {
	return func(o *ConfigOptions) {
		o.LocalPath = path
	}
}

// Remote 远端配置
func (Option) Remote(provider, endpoint, path string) option.Setting[ConfigOptions] {
	return func(o *ConfigOptions) {
		o.RemoteProvider = provider
		o.RemoteEndpoint = endpoint
		o.RemotePath = path
	}
}

// AutoUpdate 远端配置类型（etcd3,consul...）
func (Option) AutoUpdate(b bool) option.Setting[ConfigOptions] {
	return func(o *ConfigOptions) {
		o.AutoUpdate = b
	}
}

// DefaultKVs 默认配置
func (Option) DefaultKVs(kvs map[string]any) option.Setting[ConfigOptions] {
	return func(o *ConfigOptions) {
		o.DefaultKVs = kvs
	}
}
