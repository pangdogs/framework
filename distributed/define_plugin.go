package distributed

import (
	"kit.golaxy.org/golaxy/define"
)

var (
	definePlugin = define.DefineServicePlugin[Distributed](newDistributed)
	// Name 插件名称
	Name = definePlugin.Name
	// Fetch 获取插件
	Fetch = definePlugin.Fetch
	// Access 访问插件
	Access = definePlugin.Access
	// Install 安装插件
	Install = definePlugin.Install
	// Uninstall 卸载插件
	Uninstall = definePlugin.Uninstall
)
