package zap_log

import (
	"git.golaxy.org/core/define"
)

var (
	self      = define.Plugin(newLogger)
	Install   = self.Install
	Uninstall = self.Uninstall
)
