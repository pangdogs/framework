package internal

import (
	"fmt"
	"kit.golaxy.org/golaxy"
	"kit.golaxy.org/golaxy/util"
)

func Call(fun func() error) (err error) {
	if fun == nil {
		return nil
	}

	defer func() {
		if panicErr := util.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", golaxy.ErrPanicked, panicErr)
		}
	}()

	return fun()
}

func CallVoid(fun func()) (err error) {
	if fun == nil {
		return
	}

	defer func() {
		if panicErr := util.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", golaxy.ErrPanicked, panicErr)
		}
	}()

	fun()

	return nil
}
