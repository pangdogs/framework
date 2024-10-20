/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package rpc

import "git.golaxy.org/core/utils/async"

func Asserts(ret async.Ret) []any {
	return Results(ret).ensure(3)
}

func AssertVoid(ret async.Ret) {
	ResultVoid(ret).ensure(3)
}

func Assert1[T1 any](ret async.Ret) T1 {
	return Result1[T1](ret).ensure(3)
}

func Assert2[T1, T2 any](ret async.Ret) (T1, T2) {
	return Result2[T1, T2](ret).ensure(3)
}

func Assert3[T1, T2, T3 any](ret async.Ret) (T1, T2, T3) {
	return Result3[T1, T2, T3](ret).ensure(3)
}

func Assert4[T1, T2, T3, T4 any](ret async.Ret) (T1, T2, T3, T4) {
	return Result4[T1, T2, T3, T4](ret).ensure(3)
}

func Assert5[T1, T2, T3, T4, T5 any](ret async.Ret) (T1, T2, T3, T4, T5) {
	return Result5[T1, T2, T3, T4, T5](ret).ensure(3)
}

func Assert6[T1, T2, T3, T4, T5, T6 any](ret async.Ret) (T1, T2, T3, T4, T5, T6) {
	return Result6[T1, T2, T3, T4, T5, T6](ret).ensure(3)
}

func Assert7[T1, T2, T3, T4, T5, T6, T7 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7) {
	return Result7[T1, T2, T3, T4, T5, T6, T7](ret).ensure(3)
}

func Assert8[T1, T2, T3, T4, T5, T6, T7, T8 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8) {
	return Result8[T1, T2, T3, T4, T5, T6, T7, T8](ret).ensure(3)
}

func Assert9[T1, T2, T3, T4, T5, T6, T7, T8, T9 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9) {
	return Result9[T1, T2, T3, T4, T5, T6, T7, T8, T9](ret).ensure(3)
}

func Assert10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10) {
	return Result10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10](ret).ensure(3)
}

func Assert11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11) {
	return Result11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11](ret).ensure(3)
}

func Assert12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12) {
	return Result12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12](ret).ensure(3)
}

func Assert13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13) {
	return Result13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13](ret).ensure(3)
}

func Assert14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14) {
	return Result14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14](ret).ensure(3)
}

func Assert15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15) {
	return Result15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15](ret).ensure(3)
}

func Assert16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16 any](ret async.Ret) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16) {
	return Result16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16](ret).ensure(3)
}
