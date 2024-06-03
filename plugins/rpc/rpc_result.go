package rpc

import (
	"errors"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/framework/net/gap/variant"
)

var (
	ErrMethodResultCountMismatch = errors.New("rpc: method result count mismatch")
	ErrMethodResultTypeMismatch  = errors.New("rpc: method result type mismatch")
)

func Results(ret async.Ret) []any {
	if !ret.OK() {
		panic(ret.Error)
	}

	if ret.Value == nil {
		return nil
	}

	retArr := ret.Value.(variant.Array)
	rets := make([]any, len(retArr))

	for i := range rets {
		rets[i] = retArr[i].Value.Indirect()
	}

	return rets
}

func ResultVoid(ret async.Ret) {
	if !ret.OK() {
		panic(ret.Error)
	}
}

func Result1[T1 any](ret async.Ret) T1 {
	if !ret.OK() {
		panic(ret.Error)
	}

	if ret.Value == nil {
		panic(ErrMethodResultCountMismatch)
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 1 {
		panic(ErrMethodResultCountMismatch)
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	return r1
}

func Result2[T1, T2 any](ret async.Ret) (T1, T2) {
	if !ret.OK() {
		panic(ret.Error)
	}

	if ret.Value == nil {
		panic(ErrMethodResultCountMismatch)
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 2 {
		panic(ErrMethodResultCountMismatch)
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	return r1, r2
}

func Result3[T1, T2, T3 any](ret async.Ret) (T1, T2, T3) {
	if !ret.OK() {
		panic(ret.Error)
	}

	if ret.Value == nil {
		panic(ErrMethodResultCountMismatch)
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 3 {
		panic(ErrMethodResultCountMismatch)
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	return r1, r2, r3
}

func Result4[T1, T2, T3, T4 any](ret async.Ret) (T1, T2, T3, T4) {
	if !ret.OK() {
		panic(ret.Error)
	}

	if ret.Value == nil {
		panic(ErrMethodResultCountMismatch)
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 4 {
		panic(ErrMethodResultCountMismatch)
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r4, ok := retArr[3].Value.Indirect().(T4)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	return r1, r2, r3, r4
}

func Result5[T1, T2, T3, T4, T5 any](ret async.Ret) (T1, T2, T3, T4, T5) {
	if !ret.OK() {
		panic(ret.Error)
	}

	if ret.Value == nil {
		panic(ErrMethodResultCountMismatch)
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 5 {
		panic(ErrMethodResultCountMismatch)
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r4, ok := retArr[3].Value.Indirect().(T4)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r5, ok := retArr[4].Value.Indirect().(T5)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	return r1, r2, r3, r4, r5
}

func Result6[T1, T2, T3, T4, T5, T6 any](ret async.Ret) (T1, T2, T3, T4, T5, T6) {
	if !ret.OK() {
		panic(ret.Error)
	}

	if ret.Value == nil {
		panic(ErrMethodResultCountMismatch)
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 6 {
		panic(ErrMethodResultCountMismatch)
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r4, ok := retArr[3].Value.Indirect().(T4)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r5, ok := retArr[4].Value.Indirect().(T5)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r6, ok := retArr[5].Value.Indirect().(T6)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	return r1, r2, r3, r4, r5, r6
}

func Result7[T1, T2, T3, T4, T5, T6, T7 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7) {
	if !ret.OK() {
		panic(ret.Error)
	}

	if ret.Value == nil {
		panic(ErrMethodResultCountMismatch)
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 7 {
		panic(ErrMethodResultCountMismatch)
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r4, ok := retArr[3].Value.Indirect().(T4)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r5, ok := retArr[4].Value.Indirect().(T5)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r6, ok := retArr[5].Value.Indirect().(T6)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r7, ok := retArr[6].Value.Indirect().(T7)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	return r1, r2, r3, r4, r5, r6, r7
}

func Result8[T1, T2, T3, T4, T5, T6, T7, T8 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8) {
	if !ret.OK() {
		panic(ret.Error)
	}

	if ret.Value == nil {
		panic(ErrMethodResultCountMismatch)
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 8 {
		panic(ErrMethodResultCountMismatch)
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r4, ok := retArr[3].Value.Indirect().(T4)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r5, ok := retArr[4].Value.Indirect().(T5)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r6, ok := retArr[5].Value.Indirect().(T6)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r7, ok := retArr[6].Value.Indirect().(T7)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r8, ok := retArr[7].Value.Indirect().(T8)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	return r1, r2, r3, r4, r5, r6, r7, r8
}

func Result9[T1, T2, T3, T4, T5, T6, T7, T8, T9 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9) {
	if !ret.OK() {
		panic(ret.Error)
	}

	if ret.Value == nil {
		panic(ErrMethodResultCountMismatch)
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 9 {
		panic(ErrMethodResultCountMismatch)
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r4, ok := retArr[3].Value.Indirect().(T4)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r5, ok := retArr[4].Value.Indirect().(T5)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r6, ok := retArr[5].Value.Indirect().(T6)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r7, ok := retArr[6].Value.Indirect().(T7)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r8, ok := retArr[7].Value.Indirect().(T8)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r9, ok := retArr[8].Value.Indirect().(T9)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	return r1, r2, r3, r4, r5, r6, r7, r8, r9
}

func Result10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10) {
	if !ret.OK() {
		panic(ret.Error)
	}

	if ret.Value == nil {
		panic(ErrMethodResultCountMismatch)
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 10 {
		panic(ErrMethodResultCountMismatch)
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r4, ok := retArr[3].Value.Indirect().(T4)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r5, ok := retArr[4].Value.Indirect().(T5)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r6, ok := retArr[5].Value.Indirect().(T6)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r7, ok := retArr[6].Value.Indirect().(T7)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r8, ok := retArr[7].Value.Indirect().(T8)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r9, ok := retArr[8].Value.Indirect().(T9)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r10, ok := retArr[9].Value.Indirect().(T10)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	return r1, r2, r3, r4, r5, r6, r7, r8, r9, r10
}

func Result11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11) {
	if !ret.OK() {
		panic(ret.Error)
	}

	if ret.Value == nil {
		panic(ErrMethodResultCountMismatch)
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 11 {
		panic(ErrMethodResultCountMismatch)
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r4, ok := retArr[3].Value.Indirect().(T4)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r5, ok := retArr[4].Value.Indirect().(T5)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r6, ok := retArr[5].Value.Indirect().(T6)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r7, ok := retArr[6].Value.Indirect().(T7)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r8, ok := retArr[7].Value.Indirect().(T8)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r9, ok := retArr[8].Value.Indirect().(T9)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r10, ok := retArr[9].Value.Indirect().(T10)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r11, ok := retArr[10].Value.Indirect().(T11)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	return r1, r2, r3, r4, r5, r6, r7, r8, r9, r10, r11
}

func Result12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12) {
	if !ret.OK() {
		panic(ret.Error)
	}

	if ret.Value == nil {
		panic(ErrMethodResultCountMismatch)
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 12 {
		panic(ErrMethodResultCountMismatch)
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r4, ok := retArr[3].Value.Indirect().(T4)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r5, ok := retArr[4].Value.Indirect().(T5)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r6, ok := retArr[5].Value.Indirect().(T6)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r7, ok := retArr[6].Value.Indirect().(T7)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r8, ok := retArr[7].Value.Indirect().(T8)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r9, ok := retArr[8].Value.Indirect().(T9)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r10, ok := retArr[9].Value.Indirect().(T10)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r11, ok := retArr[10].Value.Indirect().(T11)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r12, ok := retArr[11].Value.Indirect().(T12)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	return r1, r2, r3, r4, r5, r6, r7, r8, r9, r10, r11, r12
}

func Result13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13) {
	if !ret.OK() {
		panic(ret.Error)
	}

	if ret.Value == nil {
		panic(ErrMethodResultCountMismatch)
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 13 {
		panic(ErrMethodResultCountMismatch)
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r4, ok := retArr[3].Value.Indirect().(T4)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r5, ok := retArr[4].Value.Indirect().(T5)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r6, ok := retArr[5].Value.Indirect().(T6)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r7, ok := retArr[6].Value.Indirect().(T7)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r8, ok := retArr[7].Value.Indirect().(T8)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r9, ok := retArr[8].Value.Indirect().(T9)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r10, ok := retArr[9].Value.Indirect().(T10)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r11, ok := retArr[10].Value.Indirect().(T11)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r12, ok := retArr[11].Value.Indirect().(T12)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r13, ok := retArr[12].Value.Indirect().(T13)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	return r1, r2, r3, r4, r5, r6, r7, r8, r9, r10, r11, r12, r13
}

func Result14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14) {
	if !ret.OK() {
		panic(ret.Error)
	}

	if ret.Value == nil {
		panic(ErrMethodResultCountMismatch)
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 14 {
		panic(ErrMethodResultCountMismatch)
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r4, ok := retArr[3].Value.Indirect().(T4)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r5, ok := retArr[4].Value.Indirect().(T5)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r6, ok := retArr[5].Value.Indirect().(T6)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r7, ok := retArr[6].Value.Indirect().(T7)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r8, ok := retArr[7].Value.Indirect().(T8)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r9, ok := retArr[8].Value.Indirect().(T9)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r10, ok := retArr[9].Value.Indirect().(T10)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r11, ok := retArr[10].Value.Indirect().(T11)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r12, ok := retArr[11].Value.Indirect().(T12)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r13, ok := retArr[12].Value.Indirect().(T13)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r14, ok := retArr[13].Value.Indirect().(T14)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	return r1, r2, r3, r4, r5, r6, r7, r8, r9, r10, r11, r12, r13, r14
}

func Result15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15) {
	if !ret.OK() {
		panic(ret.Error)
	}

	if ret.Value == nil {
		panic(ErrMethodResultCountMismatch)
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 15 {
		panic(ErrMethodResultCountMismatch)
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r4, ok := retArr[3].Value.Indirect().(T4)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r5, ok := retArr[4].Value.Indirect().(T5)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r6, ok := retArr[5].Value.Indirect().(T6)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r7, ok := retArr[6].Value.Indirect().(T7)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r8, ok := retArr[7].Value.Indirect().(T8)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r9, ok := retArr[8].Value.Indirect().(T9)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r10, ok := retArr[9].Value.Indirect().(T10)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r11, ok := retArr[10].Value.Indirect().(T11)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r12, ok := retArr[11].Value.Indirect().(T12)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r13, ok := retArr[12].Value.Indirect().(T13)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r14, ok := retArr[13].Value.Indirect().(T14)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r15, ok := retArr[14].Value.Indirect().(T15)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	return r1, r2, r3, r4, r5, r6, r7, r8, r9, r10, r11, r12, r13, r14, r15
}

func Result16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16) {
	if !ret.OK() {
		panic(ret.Error)
	}

	if ret.Value == nil {
		panic(ErrMethodResultCountMismatch)
	}

	retArr := ret.Value.(variant.Array)
	if len(retArr) < 16 {
		panic(ErrMethodResultCountMismatch)
	}

	r1, ok := retArr[0].Value.Indirect().(T1)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r2, ok := retArr[1].Value.Indirect().(T2)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r3, ok := retArr[2].Value.Indirect().(T3)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r4, ok := retArr[3].Value.Indirect().(T4)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r5, ok := retArr[4].Value.Indirect().(T5)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r6, ok := retArr[5].Value.Indirect().(T6)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r7, ok := retArr[6].Value.Indirect().(T7)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r8, ok := retArr[7].Value.Indirect().(T8)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r9, ok := retArr[8].Value.Indirect().(T9)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r10, ok := retArr[9].Value.Indirect().(T10)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r11, ok := retArr[10].Value.Indirect().(T11)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r12, ok := retArr[11].Value.Indirect().(T12)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r13, ok := retArr[12].Value.Indirect().(T13)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r14, ok := retArr[13].Value.Indirect().(T14)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r15, ok := retArr[14].Value.Indirect().(T15)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	r16, ok := retArr[15].Value.Indirect().(T16)
	if !ok {
		panic(ErrMethodResultTypeMismatch)
	}

	return r1, r2, r3, r4, r5, r6, r7, r8, r9, r10, r11, r12, r13, r14, r15, r16
}
