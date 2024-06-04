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

func parseRV[T any](retArr variant.Array, idx int) T {
	t := retArr[idx].Value.Indirect()
	r, ok := t.(T)
	if !ok && t != nil {
		panic(ErrMethodResultTypeMismatch)
	}
	return r
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

	r1 := parseRV[T1](retArr, 0)

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

	r1 := parseRV[T1](retArr, 0)
	r2 := parseRV[T2](retArr, 1)

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

	r1 := parseRV[T1](retArr, 0)
	r2 := parseRV[T2](retArr, 1)
	r3 := parseRV[T3](retArr, 2)

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

	r1 := parseRV[T1](retArr, 0)
	r2 := parseRV[T2](retArr, 1)
	r3 := parseRV[T3](retArr, 2)
	r4 := parseRV[T4](retArr, 3)

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

	r1 := parseRV[T1](retArr, 0)
	r2 := parseRV[T2](retArr, 1)
	r3 := parseRV[T3](retArr, 2)
	r4 := parseRV[T4](retArr, 3)
	r5 := parseRV[T5](retArr, 4)

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

	r1 := parseRV[T1](retArr, 0)
	r2 := parseRV[T2](retArr, 1)
	r3 := parseRV[T3](retArr, 2)
	r4 := parseRV[T4](retArr, 3)
	r5 := parseRV[T5](retArr, 4)
	r6 := parseRV[T6](retArr, 5)

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

	r1 := parseRV[T1](retArr, 0)
	r2 := parseRV[T2](retArr, 1)
	r3 := parseRV[T3](retArr, 2)
	r4 := parseRV[T4](retArr, 3)
	r5 := parseRV[T5](retArr, 4)
	r6 := parseRV[T6](retArr, 5)
	r7 := parseRV[T7](retArr, 6)

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

	r1 := parseRV[T1](retArr, 0)
	r2 := parseRV[T2](retArr, 1)
	r3 := parseRV[T3](retArr, 2)
	r4 := parseRV[T4](retArr, 3)
	r5 := parseRV[T5](retArr, 4)
	r6 := parseRV[T6](retArr, 5)
	r7 := parseRV[T7](retArr, 6)
	r8 := parseRV[T8](retArr, 7)

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

	r1 := parseRV[T1](retArr, 0)
	r2 := parseRV[T2](retArr, 1)
	r3 := parseRV[T3](retArr, 2)
	r4 := parseRV[T4](retArr, 3)
	r5 := parseRV[T5](retArr, 4)
	r6 := parseRV[T6](retArr, 5)
	r7 := parseRV[T7](retArr, 6)
	r8 := parseRV[T8](retArr, 7)
	r9 := parseRV[T9](retArr, 8)

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

	r1 := parseRV[T1](retArr, 0)
	r2 := parseRV[T2](retArr, 1)
	r3 := parseRV[T3](retArr, 2)
	r4 := parseRV[T4](retArr, 3)
	r5 := parseRV[T5](retArr, 4)
	r6 := parseRV[T6](retArr, 5)
	r7 := parseRV[T7](retArr, 6)
	r8 := parseRV[T8](retArr, 7)
	r9 := parseRV[T9](retArr, 8)
	r10 := parseRV[T10](retArr, 9)

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

	r1 := parseRV[T1](retArr, 0)
	r2 := parseRV[T2](retArr, 1)
	r3 := parseRV[T3](retArr, 2)
	r4 := parseRV[T4](retArr, 3)
	r5 := parseRV[T5](retArr, 4)
	r6 := parseRV[T6](retArr, 5)
	r7 := parseRV[T7](retArr, 6)
	r8 := parseRV[T8](retArr, 7)
	r9 := parseRV[T9](retArr, 8)
	r10 := parseRV[T10](retArr, 9)
	r11 := parseRV[T11](retArr, 10)

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

	r1 := parseRV[T1](retArr, 0)
	r2 := parseRV[T2](retArr, 1)
	r3 := parseRV[T3](retArr, 2)
	r4 := parseRV[T4](retArr, 3)
	r5 := parseRV[T5](retArr, 4)
	r6 := parseRV[T6](retArr, 5)
	r7 := parseRV[T7](retArr, 6)
	r8 := parseRV[T8](retArr, 7)
	r9 := parseRV[T9](retArr, 8)
	r10 := parseRV[T10](retArr, 9)
	r11 := parseRV[T11](retArr, 10)
	r12 := parseRV[T12](retArr, 11)

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

	r1 := parseRV[T1](retArr, 0)
	r2 := parseRV[T2](retArr, 1)
	r3 := parseRV[T3](retArr, 2)
	r4 := parseRV[T4](retArr, 3)
	r5 := parseRV[T5](retArr, 4)
	r6 := parseRV[T6](retArr, 5)
	r7 := parseRV[T7](retArr, 6)
	r8 := parseRV[T8](retArr, 7)
	r9 := parseRV[T9](retArr, 8)
	r10 := parseRV[T10](retArr, 9)
	r11 := parseRV[T11](retArr, 10)
	r12 := parseRV[T12](retArr, 11)
	r13 := parseRV[T13](retArr, 12)

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

	r1 := parseRV[T1](retArr, 0)
	r2 := parseRV[T2](retArr, 1)
	r3 := parseRV[T3](retArr, 2)
	r4 := parseRV[T4](retArr, 3)
	r5 := parseRV[T5](retArr, 4)
	r6 := parseRV[T6](retArr, 5)
	r7 := parseRV[T7](retArr, 6)
	r8 := parseRV[T8](retArr, 7)
	r9 := parseRV[T9](retArr, 8)
	r10 := parseRV[T10](retArr, 9)
	r11 := parseRV[T11](retArr, 10)
	r12 := parseRV[T12](retArr, 11)
	r13 := parseRV[T13](retArr, 12)
	r14 := parseRV[T14](retArr, 13)

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

	r1 := parseRV[T1](retArr, 0)
	r2 := parseRV[T2](retArr, 1)
	r3 := parseRV[T3](retArr, 2)
	r4 := parseRV[T4](retArr, 3)
	r5 := parseRV[T5](retArr, 4)
	r6 := parseRV[T6](retArr, 5)
	r7 := parseRV[T7](retArr, 6)
	r8 := parseRV[T8](retArr, 7)
	r9 := parseRV[T9](retArr, 8)
	r10 := parseRV[T10](retArr, 9)
	r11 := parseRV[T11](retArr, 10)
	r12 := parseRV[T12](retArr, 11)
	r13 := parseRV[T13](retArr, 12)
	r14 := parseRV[T14](retArr, 13)
	r15 := parseRV[T15](retArr, 14)

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

	r1 := parseRV[T1](retArr, 0)
	r2 := parseRV[T2](retArr, 1)
	r3 := parseRV[T3](retArr, 2)
	r4 := parseRV[T4](retArr, 3)
	r5 := parseRV[T5](retArr, 4)
	r6 := parseRV[T6](retArr, 5)
	r7 := parseRV[T7](retArr, 6)
	r8 := parseRV[T8](retArr, 7)
	r9 := parseRV[T9](retArr, 8)
	r10 := parseRV[T10](retArr, 9)
	r11 := parseRV[T11](retArr, 10)
	r12 := parseRV[T12](retArr, 11)
	r13 := parseRV[T13](retArr, 12)
	r14 := parseRV[T14](retArr, 13)
	r15 := parseRV[T15](retArr, 14)
	r16 := parseRV[T16](retArr, 15)

	return r1, r2, r3, r4, r5, r6, r7, r8, r9, r10, r11, r12, r13, r14, r15, r16
}
