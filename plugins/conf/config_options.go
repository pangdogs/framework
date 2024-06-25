package conf

import (
	"git.golaxy.org/core/utils/option"
	"github.com/spf13/viper"
)

// ConfigOptions 所有选项
type ConfigOptions struct {
	Format         string         // 配置格式（json,yaml,ini...）
	LocalPath      string         // 本地配置文件路径
	RemoteProvider string         // 远端配置类型（etcd3,consul...）
	RemoteEndpoint string         // 远端地址
	RemotePath     string         // 远端路径
	AutoUpdate     bool           // 是否热更新
	DefaultKVs     map[string]any // 默认配置
	MergeEnv       bool           // 合并环境变量
	MergeConf      *viper.Viper   // 合并配置
}

var With _Option

type _Option struct{}

// Default 默认值
func (_Option) Default() option.Setting[ConfigOptions] {
	return func(options *ConfigOptions) {
		With.Format("json").Apply(options)
		With.LocalPath("").Apply(options)
		With.Remote("", "", "").Apply(options)
		With.AutoUpdate(false).Apply(options)
		With.DefaultKVs(nil).Apply(options)
		With.MergeEnv(false).Apply(options)
		With.MergeConf(nil).Apply(options)
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

// MergeEnv 合并环境变量
func (_Option) MergeEnv(b bool) option.Setting[ConfigOptions] {
	return func(o *ConfigOptions) {
		o.MergeEnv = b
	}
}

// MergeConf 合并配置
func (_Option) MergeConf(conf *viper.Viper) option.Setting[ConfigOptions] {
	return func(o *ConfigOptions) {
		o.MergeConf = conf
	}
}
