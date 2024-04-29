package discovery

import (
	"git.golaxy.org/core/define"
)

var (
	self  = define.ServicePluginInterface[IRegistry]()
	Name  = self.Name
	Using = self.Using
)
