package conf

import (
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/framework/plugins/log"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

// IConfig 配置接口
type IConfig interface {
	IVisitConf
	Whole() IVisitConf
}

func newConfig(settings ...option.Setting[ConfigOptions]) IConfig {
	return &_Config{
		options: option.Make(With.Default(), settings...),
	}
}

type _Config struct {
	options ConfigOptions
	IVisitConf
	whole IVisitConf
}

// InitSP 初始化服务插件
func (c *_Config) InitSP(ctx service.Context) {
	log.Infof(ctx, "init plugin %q", self.Name)

	vp := viper.New()
	vp.SetConfigType(c.options.Format)

	for k, v := range c.options.DefaultKVs {
		vp.SetDefault(k, v)
	}

	local := c.options.LocalPath != ""
	remote := c.options.RemoteProvider != ""

	if local {
		vp.SetConfigFile(c.options.LocalPath)

		if err := vp.ReadInConfig(); err != nil {
			log.Panicf(ctx, "read local config %q failed, %s", c.options.LocalPath, err)
		}
	}

	if remote {
		if err := vp.AddRemoteProvider(c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath); err != nil {
			log.Panicf(ctx, "set remote config [%q, %q, %q] failed, %s", c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath, err)
		}
		if err := vp.ReadRemoteConfig(); err != nil {
			log.Panicf(ctx, "read remote config [%q, %q, %q] failed, %s", c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath, err)
		}
	}

	if c.options.AutoUpdate {
		if local {
			vp.WatchConfig()
		}
		if remote {
			if err := vp.WatchRemoteConfigOnChannel(); err != nil {
				log.Panicf(ctx, "watch remote config [%q, %q, %q] changes failed, %s", c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath, err)
			}
		}
	}

	c.IVisitConf = &_VisitConf{
		Viper: vp.Sub(ctx.GetName()),
	}
	c.whole = &_VisitConf{
		Viper: vp,
	}
}

// ShutSP 关闭服务插件
func (c *_Config) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut plugin %q", self.Name)
}

// InitRP 初始化运行时插件
func (c *_Config) InitRP(ctx runtime.Context) {
	log.Infof(ctx, "init plugin %q", self.Name)
}

// ShutRP 关闭运行时插件
func (c *_Config) ShutRP(ctx runtime.Context) {
	log.Infof(ctx, "shut plugin %q", self.Name)
}

func (c *_Config) Whole() IVisitConf {
	return c.whole
}
