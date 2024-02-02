package broker

import "git.golaxy.org/core/define"

var (
	self  = define.DefineServicePluginInterface[IBroker]()
	Name  = self.Name
	Using = self.Using
)
