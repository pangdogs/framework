package discovery

import (
	"git.golaxy.org/core/define"
)

var (
	self  = define.DefineServicePluginInterface[IRegistry]()
	Name  = self.Name
	Using = self.Using
)
