package broker

import "git.golaxy.org/core/define"

var (
	self  = define.ServicePluginInterface[IBroker]()
	Name  = self.Name
	Using = self.Using
)
