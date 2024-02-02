package conf

import (
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/framework/plugins/log"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

// IConfig 配置接口
type IConfig interface {
	All() *viper.Viper
	Service() *viper.Viper
}

func newConfig(settings ...option.Setting[ConfigOptions]) IConfig {
	return &_Config{
		options: option.Make(Option{}.Default(), settings...),
	}
}

type _Config struct {
	options           ConfigOptions
	allConf, servConf *viper.Viper
}

// InitSP 初始化服务插件
func (c *_Config) InitSP(ctx service.Context) {
	log.Infof(ctx, "init plugin %q", self.Name)

	c.allConf = viper.New()
	c.servConf = c.allConf.Sub(ctx.GetName())
}

// ShutSP 关闭服务插件
func (c *_Config) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut plugin %q", self.Name)

}

func (c *_Config) All() *viper.Viper {
	return c.allConf
}

func (c *_Config) Service() *viper.Viper {
	return c.servConf
}
