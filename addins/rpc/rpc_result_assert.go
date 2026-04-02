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

func Asserts(future async.Future) []any {
	return Results(future).ensure(3)
}

func AssertVoid(future async.Future) {
	ResultVoid(future).ensure(3)
}

func Assert1[T1 any](future async.Future) T1 {
	return Result1[T1](future).ensure(3)
}

func Assert2[T1, T2 any](future async.Future) (T1, T2) {
	return Result2[T1, T2](future).ensure(3)
}

func Assert3[T1, T2, T3 any](future async.Future) (T1, T2, T3) {
	return Result3[T1, T2, T3](future).ensure(3)
}

func Assert4[T1, T2, T3, T4 any](future async.Future) (T1, T2, T3, T4) {
	return Result4[T1, T2, T3, T4](future).ensure(3)
}

func Assert5[T1, T2, T3, T4, T5 any](future async.Future) (T1, T2, T3, T4, T5) {
	return Result5[T1, T2, T3, T4, T5](future).ensure(3)
}

func Assert6[T1, T2, T3, T4, T5, T6 any](future async.Future) (T1, T2, T3, T4, T5, T6) {
	return Result6[T1, T2, T3, T4, T5, T6](future).ensure(3)
}

func Assert7[T1, T2, T3, T4, T5, T6, T7 any](future async.Future) (T1, T2, T3, T4, T5, T6, T7) {
	return Result7[T1, T2, T3, T4, T5, T6, T7](future).ensure(3)
}

func Assert8[T1, T2, T3, T4, T5, T6, T7, T8 any](future async.Future) (T1, T2, T3, T4, T5, T6, T7, T8) {
	return Result8[T1, T2, T3, T4, T5, T6, T7, T8](future).ensure(3)
}

func Assert9[T1, T2, T3, T4, T5, T6, T7, T8, T9 any](future async.Future) (T1, T2, T3, T4, T5, T6, T7, T8, T9) {
	return Result9[T1, T2, T3, T4, T5, T6, T7, T8, T9](future).ensure(3)
}

func Assert10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10 any](future async.Future) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10) {
	return Result10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10](future).ensure(3)
}

func Assert11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11 any](future async.Future) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11) {
	return Result11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11](future).ensure(3)
}

func Assert12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12 any](future async.Future) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12) {
	return Result12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12](future).ensure(3)
}

func Assert13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13 any](future async.Future) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13) {
	return Result13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13](future).ensure(3)
}

func Assert14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14 any](future async.Future) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14) {
	return Result14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14](future).ensure(3)
}

func Assert15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15 any](future async.Future) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15) {
	return Result15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15](future).ensure(3)
}

func Assert16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16 any](future async.Future) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16) {
	return Result16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16](future).ensure(3)
}
