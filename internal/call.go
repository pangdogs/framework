package internal

import "kit.golaxy.org/golaxy/util"

func Call(fun func() error) (err error) {
	if fun == nil {
		return nil
	}

	defer func() {
		if panicErr := util.Panic2Err(); panicErr != nil {
			err = panicErr
		}
	}()

	return fun()
}

func CallVoid(fun func()) (err error) {
	if fun == nil {
		return
	}

	defer func() {
		if panicErr := util.Panic2Err(); panicErr != nil {
			err = panicErr
		}
	}()

	fun()

	return nil
}
