package dsync

import "git.golaxy.org/core/define"

var (
	self  = define.DefineServicePluginInterface[IDistSync]()
	Name  = self.Name
	Using = self.Using
)
