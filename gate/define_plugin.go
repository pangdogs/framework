package gate

import "kit.golaxy.org/golaxy/define"

var (
	definePlugin = define.DefineServicePluginInterface[Gate]()
	// Name 插件名称
	Name = definePlugin.Name
	// Get 获取插件
	Get = definePlugin.Get
	// TryGet 尝试获取插件
	TryGet = definePlugin.TryGet
)
