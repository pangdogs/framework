package log

import "git.golaxy.org/core/define"

var (
	self  = define.PluginInterface[ILogger]()
	Name  = self.Name
	Using = self.Using
)
