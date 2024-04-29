package dsync

import "git.golaxy.org/core/define"

var (
	self  = define.ServicePluginInterface[IDistSync]()
	Name  = self.Name
	Using = self.Using
)
