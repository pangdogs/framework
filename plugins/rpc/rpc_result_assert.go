package rpc

import "git.golaxy.org/core/utils/async"

func AssertResults(ret async.Ret) []any {
	return Results(ret).Ensure()
}

func AssertResultVoid(ret async.Ret) {
	ResultVoid(ret).Ensure()
}

func AssertResult1[T1 any](ret async.Ret) T1 {
	return Result1[T1](ret).Ensure()
}

func AssertResult2[T1, T2 any](ret async.Ret) (T1, T2) {
	return Result2[T1, T2](ret).Ensure()
}

func AssertResult3[T1, T2, T3 any](ret async.Ret) (T1, T2, T3) {
	return Result3[T1, T2, T3](ret).Ensure()
}

func AssertResult4[T1, T2, T3, T4 any](ret async.Ret) (T1, T2, T3, T4) {
	return Result4[T1, T2, T3, T4](ret).Ensure()
}

func AssertResult5[T1, T2, T3, T4, T5 any](ret async.Ret) (T1, T2, T3, T4, T5) {
	return Result5[T1, T2, T3, T4, T5](ret).Ensure()
}

func AssertResult6[T1, T2, T3, T4, T5, T6 any](ret async.Ret) (T1, T2, T3, T4, T5, T6) {
	return Result6[T1, T2, T3, T4, T5, T6](ret).Ensure()
}

func AssertResult7[T1, T2, T3, T4, T5, T6, T7 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7) {
	return Result7[T1, T2, T3, T4, T5, T6, T7](ret).Ensure()
}

func AssertResult8[T1, T2, T3, T4, T5, T6, T7, T8 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8) {
	return Result8[T1, T2, T3, T4, T5, T6, T7, T8](ret).Ensure()
}

func AssertResult9[T1, T2, T3, T4, T5, T6, T7, T8, T9 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9) {
	return Result9[T1, T2, T3, T4, T5, T6, T7, T8, T9](ret).Ensure()
}

func AssertResult10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10) {
	return Result10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10](ret).Ensure()
}

func AssertResult11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11) {
	return Result11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11](ret).Ensure()
}

func AssertResult12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12) {
	return Result12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12](ret).Ensure()
}

func AssertResult13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13) {
	return Result13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13](ret).Ensure()
}

func AssertResult14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14) {
	return Result14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14](ret).Ensure()
}

func AssertResult15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15) {
	return Result15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15](ret).Ensure()
}

func AssertResult16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16) {
	return Result16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16](ret).Ensure()
}
