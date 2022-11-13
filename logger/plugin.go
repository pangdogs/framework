package logger

import "github.com/galaxy-kit/galaxy-go/define"

// Plugin 定义本插件接口
var Plugin = define.DefinePluginInterface[Logger]().ServicePluginInterface()
