package rpc

import (
	"errors"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/util/types"
	"git.golaxy.org/framework/net/gap/variant"
)

var (
	ErrMethodResultCountMismatch = errors.New("rpc: method result count mismatch")
	ErrMethodResultTypeMismatch  = errors.New("rpc: method result type mismatch")
)

func Results(ret runtime.Ret) ([]any, error) {
	if !ret.OK() {
		return nil, ret.Error
	}

	if ret.Value == nil {
		return nil, nil
	}

	retArr := ret.Value.(variant.Array)
	rets := make([]any, len(retArr))

	for i := range rets {
		rets[i] = retArr[i].Value.Indirect()
	}

	return rets, nil
}

func ResultVoid(ret runtime.Ret) error {
	if !ret.OK() {
		return ret.Error
	}
	return nil
}

func Result1[T1 any](ret runtime.Ret) (T1, error) {
	if !ret.OK() {
		return types.Zero[T1](), ret.Error
	}

	if ret.Value == nil {
		return types.Zero[T1](), ErrMethodResultCountMismatch
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 1 {
		return types.Zero[T1](), ErrMethodResultCountMismatch
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		return types.Zero[T1](), ErrMethodResultTypeMismatch
	}

	return r1, nil
}

func Result2[T1, T2 any](ret runtime.Ret) (T1, T2, error) {
	if !ret.OK() {
		return types.Zero[T1](), types.Zero[T2](), ret.Error
	}

	if ret.Value == nil {
		return types.Zero[T1](), types.Zero[T2](), ErrMethodResultCountMismatch
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 2 {
		return types.Zero[T1](), types.Zero[T2](), ErrMethodResultCountMismatch
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), ErrMethodResultTypeMismatch
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), ErrMethodResultTypeMismatch
	}

	return r1, r2, nil
}

func Result3[T1, T2, T3 any](ret runtime.Ret) (T1, T2, T3, error) {
	if !ret.OK() {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), ret.Error
	}

	if ret.Value == nil {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), ErrMethodResultCountMismatch
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 3 {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), ErrMethodResultCountMismatch
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), ErrMethodResultTypeMismatch
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), ErrMethodResultTypeMismatch
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), ErrMethodResultTypeMismatch
	}

	return r1, r2, r3, nil
}

func Result4[T1, T2, T3, T4 any](ret runtime.Ret) (T1, T2, T3, T4, error) {
	if !ret.OK() {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), ret.Error
	}

	if ret.Value == nil {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), ErrMethodResultCountMismatch
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 4 {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), ErrMethodResultCountMismatch
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), ErrMethodResultTypeMismatch
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), ErrMethodResultTypeMismatch
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), ErrMethodResultTypeMismatch
	}

	r4, ok := retArr[3].Value.Indirect().(T4)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), ErrMethodResultTypeMismatch
	}

	return r1, r2, r3, r4, nil
}

func Result5[T1, T2, T3, T4, T5 any](ret runtime.Ret) (T1, T2, T3, T4, T5, error) {
	if !ret.OK() {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), ret.Error
	}

	if ret.Value == nil {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), ErrMethodResultCountMismatch
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 5 {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), ErrMethodResultCountMismatch
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), ErrMethodResultTypeMismatch
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), ErrMethodResultTypeMismatch
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), ErrMethodResultTypeMismatch
	}

	r4, ok := retArr[3].Value.Indirect().(T4)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), ErrMethodResultTypeMismatch
	}

	r5, ok := retArr[4].Value.Indirect().(T5)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), ErrMethodResultTypeMismatch
	}

	return r1, r2, r3, r4, r5, nil
}

func Result6[T1, T2, T3, T4, T5, T6 any](ret runtime.Ret) (T1, T2, T3, T4, T5, T6, error) {
	if !ret.OK() {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](), ret.Error
	}

	if ret.Value == nil {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](), ErrMethodResultCountMismatch
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 6 {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](), ErrMethodResultCountMismatch
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](), ErrMethodResultTypeMismatch
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](), ErrMethodResultTypeMismatch
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](), ErrMethodResultTypeMismatch
	}

	r4, ok := retArr[3].Value.Indirect().(T4)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](), ErrMethodResultTypeMismatch
	}

	r5, ok := retArr[4].Value.Indirect().(T5)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](), ErrMethodResultTypeMismatch
	}

	r6, ok := retArr[5].Value.Indirect().(T6)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](), ErrMethodResultTypeMismatch
	}

	return r1, r2, r3, r4, r5, r6, nil
}

func Result7[T1, T2, T3, T4, T5, T6, T7 any](ret runtime.Ret) (T1, T2, T3, T4, T5, T6, T7, error) {
	if !ret.OK() {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), ret.Error
	}

	if ret.Value == nil {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), ErrMethodResultCountMismatch
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 7 {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), ErrMethodResultCountMismatch
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), ErrMethodResultTypeMismatch
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), ErrMethodResultTypeMismatch
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), ErrMethodResultTypeMismatch
	}

	r4, ok := retArr[3].Value.Indirect().(T4)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), ErrMethodResultTypeMismatch
	}

	r5, ok := retArr[4].Value.Indirect().(T5)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), ErrMethodResultTypeMismatch
	}

	r6, ok := retArr[5].Value.Indirect().(T6)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), ErrMethodResultTypeMismatch
	}

	r7, ok := retArr[6].Value.Indirect().(T7)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), ErrMethodResultTypeMismatch
	}

	return r1, r2, r3, r4, r5, r6, r7, nil
}

func Result8[T1, T2, T3, T4, T5, T6, T7, T8 any](ret runtime.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, error) {
	if !ret.OK() {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), ret.Error
	}

	if ret.Value == nil {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), ErrMethodResultCountMismatch
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 8 {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), ErrMethodResultCountMismatch
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), ErrMethodResultTypeMismatch
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), ErrMethodResultTypeMismatch
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), ErrMethodResultTypeMismatch
	}

	r4, ok := retArr[3].Value.Indirect().(T4)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), ErrMethodResultTypeMismatch
	}

	r5, ok := retArr[4].Value.Indirect().(T5)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), ErrMethodResultTypeMismatch
	}

	r6, ok := retArr[5].Value.Indirect().(T6)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), ErrMethodResultTypeMismatch
	}

	r7, ok := retArr[6].Value.Indirect().(T7)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), ErrMethodResultTypeMismatch
	}

	r8, ok := retArr[7].Value.Indirect().(T8)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), ErrMethodResultTypeMismatch
	}

	return r1, r2, r3, r4, r5, r6, r7, r8, nil
}

func Result9[T1, T2, T3, T4, T5, T6, T7, T8, T9 any](ret runtime.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, error) {
	if !ret.OK() {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), ret.Error
	}

	if ret.Value == nil {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), ErrMethodResultCountMismatch
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 9 {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), ErrMethodResultCountMismatch
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), ErrMethodResultTypeMismatch
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), ErrMethodResultTypeMismatch
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), ErrMethodResultTypeMismatch
	}

	r4, ok := retArr[3].Value.Indirect().(T4)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), ErrMethodResultTypeMismatch
	}

	r5, ok := retArr[4].Value.Indirect().(T5)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), ErrMethodResultTypeMismatch
	}

	r6, ok := retArr[5].Value.Indirect().(T6)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), ErrMethodResultTypeMismatch
	}

	r7, ok := retArr[6].Value.Indirect().(T7)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), ErrMethodResultTypeMismatch
	}

	r8, ok := retArr[7].Value.Indirect().(T8)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), ErrMethodResultTypeMismatch
	}

	r9, ok := retArr[8].Value.Indirect().(T9)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), ErrMethodResultTypeMismatch
	}

	return r1, r2, r3, r4, r5, r6, r7, r8, r9, nil
}

func Result10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10 any](ret runtime.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, error) {
	if !ret.OK() {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), ret.Error
	}

	if ret.Value == nil {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), ErrMethodResultCountMismatch
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 10 {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), ErrMethodResultCountMismatch
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), ErrMethodResultTypeMismatch
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), ErrMethodResultTypeMismatch
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), ErrMethodResultTypeMismatch
	}

	r4, ok := retArr[3].Value.Indirect().(T4)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), ErrMethodResultTypeMismatch
	}

	r5, ok := retArr[4].Value.Indirect().(T5)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), ErrMethodResultTypeMismatch
	}

	r6, ok := retArr[5].Value.Indirect().(T6)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), ErrMethodResultTypeMismatch
	}

	r7, ok := retArr[6].Value.Indirect().(T7)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), ErrMethodResultTypeMismatch
	}

	r8, ok := retArr[7].Value.Indirect().(T8)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), ErrMethodResultTypeMismatch
	}

	r9, ok := retArr[8].Value.Indirect().(T9)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), ErrMethodResultTypeMismatch
	}

	r10, ok := retArr[9].Value.Indirect().(T10)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), ErrMethodResultTypeMismatch
	}

	return r1, r2, r3, r4, r5, r6, r7, r8, r9, r10, nil
}

func Result11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11 any](ret runtime.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, error) {
	if !ret.OK() {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), ret.Error
	}

	if ret.Value == nil {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), ErrMethodResultCountMismatch
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 11 {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), ErrMethodResultCountMismatch
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), ErrMethodResultTypeMismatch
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), ErrMethodResultTypeMismatch
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), ErrMethodResultTypeMismatch
	}

	r4, ok := retArr[3].Value.Indirect().(T4)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), ErrMethodResultTypeMismatch
	}

	r5, ok := retArr[4].Value.Indirect().(T5)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), ErrMethodResultTypeMismatch
	}

	r6, ok := retArr[5].Value.Indirect().(T6)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), ErrMethodResultTypeMismatch
	}

	r7, ok := retArr[6].Value.Indirect().(T7)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), ErrMethodResultTypeMismatch
	}

	r8, ok := retArr[7].Value.Indirect().(T8)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), ErrMethodResultTypeMismatch
	}

	r9, ok := retArr[8].Value.Indirect().(T9)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), ErrMethodResultTypeMismatch
	}

	r10, ok := retArr[9].Value.Indirect().(T10)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), ErrMethodResultTypeMismatch
	}

	r11, ok := retArr[10].Value.Indirect().(T11)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), ErrMethodResultTypeMismatch
	}

	return r1, r2, r3, r4, r5, r6, r7, r8, r9, r10, r11, nil
}

func Result12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12 any](ret runtime.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, error) {
	if !ret.OK() {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](), ret.Error
	}

	if ret.Value == nil {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](), ErrMethodResultCountMismatch
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 12 {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](), ErrMethodResultCountMismatch
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](), ErrMethodResultTypeMismatch
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](), ErrMethodResultTypeMismatch
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](), ErrMethodResultTypeMismatch
	}

	r4, ok := retArr[3].Value.Indirect().(T4)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](), ErrMethodResultTypeMismatch
	}

	r5, ok := retArr[4].Value.Indirect().(T5)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](), ErrMethodResultTypeMismatch
	}

	r6, ok := retArr[5].Value.Indirect().(T6)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](), ErrMethodResultTypeMismatch
	}

	r7, ok := retArr[6].Value.Indirect().(T7)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](), ErrMethodResultTypeMismatch
	}

	r8, ok := retArr[7].Value.Indirect().(T8)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](), ErrMethodResultTypeMismatch
	}

	r9, ok := retArr[8].Value.Indirect().(T9)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](), ErrMethodResultTypeMismatch
	}

	r10, ok := retArr[9].Value.Indirect().(T10)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](), ErrMethodResultTypeMismatch
	}

	r11, ok := retArr[10].Value.Indirect().(T11)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](), ErrMethodResultTypeMismatch
	}

	r12, ok := retArr[11].Value.Indirect().(T12)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](), ErrMethodResultTypeMismatch
	}

	return r1, r2, r3, r4, r5, r6, r7, r8, r9, r10, r11, r12, nil
}

func Result13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13 any](ret runtime.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, error) {
	if !ret.OK() {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), ret.Error
	}

	if ret.Value == nil {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), ErrMethodResultCountMismatch
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 13 {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), ErrMethodResultCountMismatch
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), ErrMethodResultCountMismatch
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), ErrMethodResultCountMismatch
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), ErrMethodResultCountMismatch
	}

	r4, ok := retArr[3].Value.Indirect().(T4)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), ErrMethodResultCountMismatch
	}

	r5, ok := retArr[4].Value.Indirect().(T5)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), ErrMethodResultCountMismatch
	}

	r6, ok := retArr[5].Value.Indirect().(T6)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), ErrMethodResultCountMismatch
	}

	r7, ok := retArr[6].Value.Indirect().(T7)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), ErrMethodResultCountMismatch
	}

	r8, ok := retArr[7].Value.Indirect().(T8)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), ErrMethodResultCountMismatch
	}

	r9, ok := retArr[8].Value.Indirect().(T9)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), ErrMethodResultCountMismatch
	}

	r10, ok := retArr[9].Value.Indirect().(T10)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), ErrMethodResultCountMismatch
	}

	r11, ok := retArr[10].Value.Indirect().(T11)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), ErrMethodResultCountMismatch
	}

	r12, ok := retArr[11].Value.Indirect().(T12)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), ErrMethodResultCountMismatch
	}

	r13, ok := retArr[12].Value.Indirect().(T13)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), ErrMethodResultCountMismatch
	}

	return r1, r2, r3, r4, r5, r6, r7, r8, r9, r10, r11, r12, r13, nil
}

func Result14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14 any](ret runtime.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, error) {
	if !ret.OK() {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), ret.Error
	}

	if ret.Value == nil {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), ErrMethodResultCountMismatch
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 14 {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), ErrMethodResultCountMismatch
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), ErrMethodResultCountMismatch
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), ErrMethodResultCountMismatch
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), ErrMethodResultCountMismatch
	}

	r4, ok := retArr[3].Value.Indirect().(T4)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), ErrMethodResultCountMismatch
	}

	r5, ok := retArr[4].Value.Indirect().(T5)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), ErrMethodResultCountMismatch
	}

	r6, ok := retArr[5].Value.Indirect().(T6)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), ErrMethodResultCountMismatch
	}

	r7, ok := retArr[6].Value.Indirect().(T7)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), ErrMethodResultCountMismatch
	}

	r8, ok := retArr[7].Value.Indirect().(T8)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), ErrMethodResultCountMismatch
	}

	r9, ok := retArr[8].Value.Indirect().(T9)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), ErrMethodResultCountMismatch
	}

	r10, ok := retArr[9].Value.Indirect().(T10)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), ErrMethodResultCountMismatch
	}

	r11, ok := retArr[10].Value.Indirect().(T11)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), ErrMethodResultCountMismatch
	}

	r12, ok := retArr[11].Value.Indirect().(T12)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), ErrMethodResultCountMismatch
	}

	r13, ok := retArr[12].Value.Indirect().(T13)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), ErrMethodResultCountMismatch
	}

	r14, ok := retArr[13].Value.Indirect().(T14)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), ErrMethodResultCountMismatch
	}

	return r1, r2, r3, r4, r5, r6, r7, r8, r9, r10, r11, r12, r13, r14, nil
}

func Result15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15 any](ret runtime.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, error) {
	if !ret.OK() {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), ret.Error
	}

	if ret.Value == nil {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), ErrMethodResultCountMismatch
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 15 {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), ErrMethodResultCountMismatch
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), ErrMethodResultCountMismatch
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), ErrMethodResultCountMismatch
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), ErrMethodResultCountMismatch
	}

	r4, ok := retArr[3].Value.Indirect().(T4)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), ErrMethodResultCountMismatch
	}

	r5, ok := retArr[4].Value.Indirect().(T5)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), ErrMethodResultCountMismatch
	}

	r6, ok := retArr[5].Value.Indirect().(T6)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), ErrMethodResultCountMismatch
	}

	r7, ok := retArr[6].Value.Indirect().(T7)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), ErrMethodResultCountMismatch
	}

	r8, ok := retArr[7].Value.Indirect().(T8)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), ErrMethodResultCountMismatch
	}

	r9, ok := retArr[8].Value.Indirect().(T9)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), ErrMethodResultCountMismatch
	}

	r10, ok := retArr[9].Value.Indirect().(T10)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), ErrMethodResultCountMismatch
	}

	r11, ok := retArr[10].Value.Indirect().(T11)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), ErrMethodResultCountMismatch
	}

	r12, ok := retArr[11].Value.Indirect().(T12)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), ErrMethodResultCountMismatch
	}

	r13, ok := retArr[12].Value.Indirect().(T13)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), ErrMethodResultCountMismatch
	}

	r14, ok := retArr[13].Value.Indirect().(T14)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), ErrMethodResultCountMismatch
	}

	r15, ok := retArr[14].Value.Indirect().(T15)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), ErrMethodResultCountMismatch
	}

	return r1, r2, r3, r4, r5, r6, r7, r8, r9, r10, r11, r12, r13, r14, r15, nil
}

func Result16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16 any](ret runtime.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16, error) {
	if !ret.OK() {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), types.Zero[T16](), ret.Error
	}

	if ret.Value == nil {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), types.Zero[T16](), ErrMethodResultCountMismatch
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 16 {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), types.Zero[T16](), ErrMethodResultCountMismatch
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), types.Zero[T16](), ErrMethodResultCountMismatch
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), types.Zero[T16](), ErrMethodResultCountMismatch
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), types.Zero[T16](), ErrMethodResultCountMismatch
	}

	r4, ok := retArr[3].Value.Indirect().(T4)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), types.Zero[T16](), ErrMethodResultCountMismatch
	}

	r5, ok := retArr[4].Value.Indirect().(T5)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), types.Zero[T16](), ErrMethodResultCountMismatch
	}

	r6, ok := retArr[5].Value.Indirect().(T6)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), types.Zero[T16](), ErrMethodResultCountMismatch
	}

	r7, ok := retArr[6].Value.Indirect().(T7)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), types.Zero[T16](), ErrMethodResultCountMismatch
	}

	r8, ok := retArr[7].Value.Indirect().(T8)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), types.Zero[T16](), ErrMethodResultCountMismatch
	}

	r9, ok := retArr[8].Value.Indirect().(T9)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), types.Zero[T16](), ErrMethodResultCountMismatch
	}

	r10, ok := retArr[9].Value.Indirect().(T10)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), types.Zero[T16](), ErrMethodResultCountMismatch
	}

	r11, ok := retArr[10].Value.Indirect().(T11)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), types.Zero[T16](), ErrMethodResultCountMismatch
	}

	r12, ok := retArr[11].Value.Indirect().(T12)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), types.Zero[T16](), ErrMethodResultCountMismatch
	}

	r13, ok := retArr[12].Value.Indirect().(T13)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), types.Zero[T16](), ErrMethodResultCountMismatch
	}

	r14, ok := retArr[13].Value.Indirect().(T14)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), types.Zero[T16](), ErrMethodResultCountMismatch
	}

	r15, ok := retArr[14].Value.Indirect().(T15)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), types.Zero[T16](), ErrMethodResultCountMismatch
	}

	r16, ok := retArr[15].Value.Indirect().(T16)
	if !ok {
		return types.Zero[T1](), types.Zero[T2](), types.Zero[T3](), types.Zero[T4](), types.Zero[T5](), types.Zero[T6](),
			types.Zero[T7](), types.Zero[T8](), types.Zero[T9](), types.Zero[T10](), types.Zero[T11](), types.Zero[T12](),
			types.Zero[T13](), types.Zero[T14](), types.Zero[T15](), types.Zero[T16](), ErrMethodResultCountMismatch
	}

	return r1, r2, r3, r4, r5, r6, r7, r8, r9, r10, r11, r12, r13, r14, r15, r16, nil
}
