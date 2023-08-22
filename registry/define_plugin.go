package registry

import (
	"kit.golaxy.org/golaxy/define"
)

var (
	definePlugin = define.DefineServicePluginInterface[Registry]()
	// Name 插件名称
	Name = definePlugin.Name
	// Fetch 获取插件
	Fetch = definePlugin.Fetch
	// Access 访问插件
	Access = definePlugin.Access
)
