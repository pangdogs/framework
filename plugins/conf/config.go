package conf

import (
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/plugins/log"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"time"
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
	*_VisitConf
	whole *_VisitConf
}

// InitSP 初始化服务插件
func (c *_Config) InitSP(ctx service.Context) {
	log.Infof(ctx, "init plugin %q", self.Name)

	vp := viper.New()
	vp.SetConfigType(c.options.Format)

	for k, v := range c.options.DefaultKVs {
		vp.SetDefault(k, v)
	}

	if c.options.AutoEnv {
		vp.AutomaticEnv()
	}

	if c.options.AutoPFlags {
		vp.BindPFlags(pflag.CommandLine)
	}

	local := c.options.LocalPath != ""
	remote := c.options.RemoteProvider != ""

	if local {
		vp.SetConfigFile(c.options.LocalPath)

		if err := vp.ReadInConfig(); err != nil {
			log.Panicf(ctx, "read local config %q failed, %s", c.options.LocalPath, err)
		}

		log.Infof(ctx, "load local config %q config ok", c.options.LocalPath)
	}

	if remote {
		if err := vp.AddRemoteProvider(c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath); err != nil {
			log.Panicf(ctx, "set remote config [%q, %q, %q] failed, %s", c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath, err)
		}
		if err := vp.ReadRemoteConfig(); err != nil {
			log.Panicf(ctx, "read remote config [%q, %q, %q] failed, %s", c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath, err)
		}

		log.Infof(ctx, "load remote config [%q, %q, %q] ok", c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath)
	}

	subVp := vp.Sub(ctx.GetName())
	if subVp == nil {
		subVp = viper.New()
	}
	c._VisitConf = &_VisitConf{
		Viper: subVp,
	}
	c.whole = &_VisitConf{
		Viper: vp,
	}

	if c.options.AutoUpdate {
		if local {
			vp.OnConfigChange(func(in fsnotify.Event) {
				subVp := vp.Sub(ctx.GetName())
				if subVp == nil {
					subVp = viper.New()
				}
				c._VisitConf.Viper = subVp

				log.Infof(ctx, "reload local config %q ok", c.options.LocalPath)
			})
			vp.WatchConfig()
		}
		if remote {
			go func() {
				for {
					time.Sleep(time.Second * 3)

					err := vp.WatchRemoteConfig()
					if err != nil {
						log.Errorf(ctx, "watch remote config [%q, %q, %q] changes failed, %s", c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath, err)
						continue
					}

					subVp := vp.Sub(ctx.GetName())
					if subVp == nil {
						subVp = viper.New()
					}
					c._VisitConf.Viper = subVp

					log.Infof(ctx, "reload remote config [%q, %q, %q] ok", c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath)
				}
			}()
		}
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
