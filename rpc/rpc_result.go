package rpc

import (
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/plugins/gap/variant"
)

func Results(ret runtime.Ret) ([]any, error) {
	if !ret.OK() {
		return nil, ret.Error
	}

	if ret.Value == nil {
		return nil, nil
	}

	rvArr := ret.Value.(variant.Array)
	rvs := make([]any, len(rvArr))

	for i := range rvs {
		rvs[i] = rvArr[i].Value.Indirect()
	}

	return rvs, nil
}

func Result(ret runtime.Ret) (any, error) {
	if !ret.OK() {
		return nil, ret.Error
	}

	if ret.Value == nil {
		return nil, nil
	}

	rvArr := ret.Value.(variant.Array)
	if len(rvArr) <= 0 {
		return nil, nil
	}

	return rvArr[0].Value.Indirect(), nil
}
