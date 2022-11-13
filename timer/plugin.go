package timer

import "github.com/galaxy-kit/galaxy-go/define"

var Plugin = define.DefinePlugin[Timer, any]().RuntimePlugin(newTimer)
