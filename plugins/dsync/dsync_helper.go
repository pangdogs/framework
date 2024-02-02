package dsync

import (
	"fmt"
	"git.golaxy.org/core"
	"strings"
)

// Path return name path.
func Path(dsync IDistSync, elems ...string) string {
	if dsync == nil {
		panic(fmt.Errorf("%w: dsync is nil", core.ErrArgs))
	}
	return strings.Join(elems, dsync.GetSeparator())
}
