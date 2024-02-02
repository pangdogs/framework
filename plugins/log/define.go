package log

import "git.golaxy.org/core/define"

var (
	self  = define.DefinePluginInterface[ILogger]()
	Name  = self.Name
	Using = self.Using
)
