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
		options: option.Make(Option{}.Default(), settings...),
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

	if c.options.AutoUpdate {
		vp.WatchConfig()
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
