package conf

import (
	"git.golaxy.org/core/util/option"
)

// ConfigOptions 所有选项
type ConfigOptions struct {
	AutoEnv        bool           // 是否合并环境变量
	AutoPFlags     bool           // 是否合并启动参数
	Format         string         // 配置格式（json,yaml,ini...）
	LocalPath      string         // 本地配置文件路径
	RemoteProvider string         // 远端配置类型（etcd3,consul...）
	RemoteEndpoint string         // 远端地址
	RemotePath     string         // 远端路径
	AutoUpdate     bool           // 是否热更新
	DefaultKVs     map[string]any // 默认配置
}

var With _Option

type _Option struct{}

// Default 默认值
func (_Option) Default() option.Setting[ConfigOptions] {
	return func(options *ConfigOptions) {
		With.AutoEnv(true).Apply(options)
		With.AutoPFlags(true).Apply(options)
		With.Format("json").Apply(options)
		With.LocalPath("").Apply(options)
		With.Remote("", "", "").Apply(options)
		With.AutoUpdate(false).Apply(options)
		With.DefaultKVs(nil).Apply(options)
	}
}

// Format 配置格式（json,yaml,ini...）
func (_Option) Format(format string) option.Setting[ConfigOptions] {
	return func(o *ConfigOptions) {
		o.Format = format
	}
}

// LocalPath 本地配置文件路径
func (_Option) LocalPath(path string) option.Setting[ConfigOptions] {
	return func(o *ConfigOptions) {
		o.LocalPath = path
	}
}

// Remote 远端配置
func (_Option) Remote(provider, endpoint, path string) option.Setting[ConfigOptions] {
	return func(o *ConfigOptions) {
		o.RemoteProvider = provider
		o.RemoteEndpoint = endpoint
		o.RemotePath = path
	}
}

// AutoUpdate 远端配置类型（etcd3,consul...）
func (_Option) AutoUpdate(b bool) option.Setting[ConfigOptions] {
	return func(o *ConfigOptions) {
		o.AutoUpdate = b
	}
}

// DefaultKVs 默认配置
func (_Option) DefaultKVs(kvs map[string]any) option.Setting[ConfigOptions] {
	return func(o *ConfigOptions) {
		o.DefaultKVs = kvs
	}
}

// AutoEnv 是否合并环境变量
func (_Option) AutoEnv(b bool) option.Setting[ConfigOptions] {
	return func(o *ConfigOptions) {
		o.AutoEnv = b
	}
}

// AutoPFlags 是否合并启动参数
func (_Option) AutoPFlags(b bool) option.Setting[ConfigOptions] {
	return func(o *ConfigOptions) {
		o.AutoPFlags = b
	}
}
