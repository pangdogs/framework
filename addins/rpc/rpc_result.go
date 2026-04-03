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

import (
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/exception"
)

type ResultValues struct {
	Values []any
	Error  error
}

func (rvs ResultValues) Extract() ([]any, error) {
	return rvs.Values, rvs.Error
}

func (rvs ResultValues) Ensure() []any {
	return rvs.ensure(2)
}

func (rvs ResultValues) ensure(skip int) []any {
	if rvs.Error != nil {
		exception.PanicSkip(skip, rvs.Error)
	}
	return rvs.Values
}

func Results(future async.Future) (rvs ResultValues) {
	return ParseResults(<-future.Chan())
}

type ResultTupleVoid struct {
	Error error
}

func (rtp ResultTupleVoid) Extract() error {
	return rtp.Error
}

func (rtp ResultTupleVoid) Ensure() {
	rtp.ensure(2)
}

func (rtp ResultTupleVoid) ensure(skip int) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
}

func ResultVoid(future async.Future) (rtp ResultTupleVoid) {
	return ParseVoid(<-future.Chan())
}

type ResultTuple1[T1 any] struct {
	R1    T1
	Error error
}

func (rtp ResultTuple1[T1]) Extract() (T1, error) {
	return rtp.R1, rtp.Error
}

func (rtp ResultTuple1[T1]) Ensure() T1 {
	return rtp.ensure(2)
}

func (rtp ResultTuple1[T1]) ensure(skip int) T1 {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1
}

func Result1[T1 any](future async.Future) (rtp ResultTuple1[T1]) {
	return Parse1[T1](<-future.Chan())
}

type ResultTuple2[T1, T2 any] struct {
	R1    T1
	R2    T2
	Error error
}

func (rtp ResultTuple2[T1, T2]) Extract() (T1, T2, error) {
	return rtp.R1, rtp.R2, rtp.Error
}

func (rtp ResultTuple2[T1, T2]) Ensure() (T1, T2) {
	return rtp.ensure(2)
}

func (rtp ResultTuple2[T1, T2]) ensure(skip int) (T1, T2) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2
}

func Result2[T1, T2 any](future async.Future) (rtp ResultTuple2[T1, T2]) {
	return Parse2[T1, T2](<-future.Chan())
}

type ResultTuple3[T1, T2, T3 any] struct {
	R1    T1
	R2    T2
	R3    T3
	Error error
}

func (rtp ResultTuple3[T1, T2, T3]) Extract() (T1, T2, T3, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.Error
}

func (rtp ResultTuple3[T1, T2, T3]) Ensure() (T1, T2, T3) {
	return rtp.ensure(2)
}

func (rtp ResultTuple3[T1, T2, T3]) ensure(skip int) (T1, T2, T3) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3
}

func Result3[T1, T2, T3 any](future async.Future) (rtp ResultTuple3[T1, T2, T3]) {
	return Parse3[T1, T2, T3](<-future.Chan())
}

type ResultTuple4[T1, T2, T3, T4 any] struct {
	R1    T1
	R2    T2
	R3    T3
	R4    T4
	Error error
}

func (rtp ResultTuple4[T1, T2, T3, T4]) Extract() (T1, T2, T3, T4, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.Error
}

func (rtp ResultTuple4[T1, T2, T3, T4]) Ensure() (T1, T2, T3, T4) {
	return rtp.ensure(2)
}

func (rtp ResultTuple4[T1, T2, T3, T4]) ensure(skip int) (T1, T2, T3, T4) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4
}

func Result4[T1, T2, T3, T4 any](future async.Future) (rtp ResultTuple4[T1, T2, T3, T4]) {
	return Parse4[T1, T2, T3, T4](<-future.Chan())
}

type ResultTuple5[T1, T2, T3, T4, T5 any] struct {
	R1    T1
	R2    T2
	R3    T3
	R4    T4
	R5    T5
	Error error
}

func (rtp ResultTuple5[T1, T2, T3, T4, T5]) Extract() (T1, T2, T3, T4, T5, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.Error
}

func (rtp ResultTuple5[T1, T2, T3, T4, T5]) Ensure() (T1, T2, T3, T4, T5) {
	return rtp.ensure(2)
}

func (rtp ResultTuple5[T1, T2, T3, T4, T5]) ensure(skip int) (T1, T2, T3, T4, T5) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5
}

func Result5[T1, T2, T3, T4, T5 any](future async.Future) (rtp ResultTuple5[T1, T2, T3, T4, T5]) {
	return Parse5[T1, T2, T3, T4, T5](<-future.Chan())
}

type ResultTuple6[T1, T2, T3, T4, T5, T6 any] struct {
	R1    T1
	R2    T2
	R3    T3
	R4    T4
	R5    T5
	R6    T6
	Error error
}

func (rtp ResultTuple6[T1, T2, T3, T4, T5, T6]) Extract() (T1, T2, T3, T4, T5, T6, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.Error
}

func (rtp ResultTuple6[T1, T2, T3, T4, T5, T6]) Ensure() (T1, T2, T3, T4, T5, T6) {
	return rtp.ensure(2)
}

func (rtp ResultTuple6[T1, T2, T3, T4, T5, T6]) ensure(skip int) (T1, T2, T3, T4, T5, T6) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6
}

func Result6[T1, T2, T3, T4, T5, T6 any](future async.Future) (rtp ResultTuple6[T1, T2, T3, T4, T5, T6]) {
	return Parse6[T1, T2, T3, T4, T5, T6](<-future.Chan())
}

type ResultTuple7[T1, T2, T3, T4, T5, T6, T7 any] struct {
	R1    T1
	R2    T2
	R3    T3
	R4    T4
	R5    T5
	R6    T6
	R7    T7
	Error error
}

func (rtp ResultTuple7[T1, T2, T3, T4, T5, T6, T7]) Extract() (T1, T2, T3, T4, T5, T6, T7, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.Error
}

func (rtp ResultTuple7[T1, T2, T3, T4, T5, T6, T7]) Ensure() (T1, T2, T3, T4, T5, T6, T7) {
	return rtp.ensure(2)
}

func (rtp ResultTuple7[T1, T2, T3, T4, T5, T6, T7]) ensure(skip int) (T1, T2, T3, T4, T5, T6, T7) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7
}

func Result7[T1, T2, T3, T4, T5, T6, T7 any](future async.Future) (rtp ResultTuple7[T1, T2, T3, T4, T5, T6, T7]) {
	return Parse7[T1, T2, T3, T4, T5, T6, T7](<-future.Chan())
}

type ResultTuple8[T1, T2, T3, T4, T5, T6, T7, T8 any] struct {
	R1    T1
	R2    T2
	R3    T3
	R4    T4
	R5    T5
	R6    T6
	R7    T7
	R8    T8
	Error error
}

func (rtp ResultTuple8[T1, T2, T3, T4, T5, T6, T7, T8]) Extract() (T1, T2, T3, T4, T5, T6, T7, T8, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.Error
}

func (rtp ResultTuple8[T1, T2, T3, T4, T5, T6, T7, T8]) Ensure() (T1, T2, T3, T4, T5, T6, T7, T8) {
	return rtp.ensure(2)
}

func (rtp ResultTuple8[T1, T2, T3, T4, T5, T6, T7, T8]) ensure(skip int) (T1, T2, T3, T4, T5, T6, T7, T8) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8
}

func Result8[T1, T2, T3, T4, T5, T6, T7, T8 any](future async.Future) (rtp ResultTuple8[T1, T2, T3, T4, T5, T6, T7, T8]) {
	return Parse8[T1, T2, T3, T4, T5, T6, T7, T8](<-future.Chan())
}

type ResultTuple9[T1, T2, T3, T4, T5, T6, T7, T8, T9 any] struct {
	R1    T1
	R2    T2
	R3    T3
	R4    T4
	R5    T5
	R6    T6
	R7    T7
	R8    T8
	R9    T9
	Error error
}

func (rtp ResultTuple9[T1, T2, T3, T4, T5, T6, T7, T8, T9]) Extract() (T1, T2, T3, T4, T5, T6, T7, T8, T9, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.Error
}

func (rtp ResultTuple9[T1, T2, T3, T4, T5, T6, T7, T8, T9]) Ensure() (T1, T2, T3, T4, T5, T6, T7, T8, T9) {
	return rtp.ensure(2)
}

func (rtp ResultTuple9[T1, T2, T3, T4, T5, T6, T7, T8, T9]) ensure(skip int) (T1, T2, T3, T4, T5, T6, T7, T8, T9) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9
}

func Result9[T1, T2, T3, T4, T5, T6, T7, T8, T9 any](future async.Future) (rtp ResultTuple9[T1, T2, T3, T4, T5, T6, T7, T8, T9]) {
	return Parse9[T1, T2, T3, T4, T5, T6, T7, T8, T9](<-future.Chan())
}

type ResultTuple10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10 any] struct {
	R1    T1
	R2    T2
	R3    T3
	R4    T4
	R5    T5
	R6    T6
	R7    T7
	R8    T8
	R9    T9
	R10   T10
	Error error
}

func (rtp ResultTuple10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) Extract() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.Error
}

func (rtp ResultTuple10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) Ensure() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10) {
	return rtp.ensure(2)
}

func (rtp ResultTuple10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) ensure(skip int) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10
}

func Result10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10 any](future async.Future) (rtp ResultTuple10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) {
	return Parse10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10](<-future.Chan())
}

type ResultTuple11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11 any] struct {
	R1    T1
	R2    T2
	R3    T3
	R4    T4
	R5    T5
	R6    T6
	R7    T7
	R8    T8
	R9    T9
	R10   T10
	R11   T11
	Error error
}

func (rtp ResultTuple11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]) Extract() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.Error
}

func (rtp ResultTuple11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]) Ensure() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11) {
	return rtp.ensure(2)
}

func (rtp ResultTuple11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]) ensure(skip int) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11
}

func Result11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11 any](future async.Future) (rtp ResultTuple11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]) {
	return Parse11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11](<-future.Chan())
}

type ResultTuple12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12 any] struct {
	R1    T1
	R2    T2
	R3    T3
	R4    T4
	R5    T5
	R6    T6
	R7    T7
	R8    T8
	R9    T9
	R10   T10
	R11   T11
	R12   T12
	Error error
}

func (rtp ResultTuple12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) Extract() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.R12, rtp.Error
}

func (rtp ResultTuple12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) Ensure() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12) {
	return rtp.ensure(2)
}

func (rtp ResultTuple12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) ensure(skip int) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.R12
}

func Result12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12 any](future async.Future) (rtp ResultTuple12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) {
	return Parse12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12](<-future.Chan())
}

type ResultTuple13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13 any] struct {
	R1    T1
	R2    T2
	R3    T3
	R4    T4
	R5    T5
	R6    T6
	R7    T7
	R8    T8
	R9    T9
	R10   T10
	R11   T11
	R12   T12
	R13   T13
	Error error
}

func (rtp ResultTuple13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) Extract() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.R12, rtp.R13, rtp.Error
}

func (rtp ResultTuple13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) Ensure() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13) {
	return rtp.ensure(2)
}

func (rtp ResultTuple13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) ensure(skip int) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.R12, rtp.R13
}

func Result13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13 any](future async.Future) (rtp ResultTuple13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) {
	return Parse13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13](<-future.Chan())
}

type ResultTuple14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14 any] struct {
	R1    T1
	R2    T2
	R3    T3
	R4    T4
	R5    T5
	R6    T6
	R7    T7
	R8    T8
	R9    T9
	R10   T10
	R11   T11
	R12   T12
	R13   T13
	R14   T14
	Error error
}

func (rtp ResultTuple14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) Extract() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.R12, rtp.R13, rtp.R14, rtp.Error
}

func (rtp ResultTuple14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) Ensure() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14) {
	return rtp.ensure(2)
}

func (rtp ResultTuple14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) ensure(skip int) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.R12, rtp.R13, rtp.R14
}

func Result14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14 any](future async.Future) (rtp ResultTuple14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) {
	return Parse14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14](<-future.Chan())
}

type ResultTuple15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15 any] struct {
	R1    T1
	R2    T2
	R3    T3
	R4    T4
	R5    T5
	R6    T6
	R7    T7
	R8    T8
	R9    T9
	R10   T10
	R11   T11
	R12   T12
	R13   T13
	R14   T14
	R15   T15
	Error error
}

func (rtp ResultTuple15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) Extract() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.R12, rtp.R13, rtp.R14, rtp.R15, rtp.Error
}

func (rtp ResultTuple15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) Ensure() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15) {
	return rtp.ensure(2)
}

func (rtp ResultTuple15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) ensure(skip int) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.R12, rtp.R13, rtp.R14, rtp.R15
}

func Result15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15 any](future async.Future) (rtp ResultTuple15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) {
	return Parse15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15](<-future.Chan())
}

type ResultTuple16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16 any] struct {
	R1    T1
	R2    T2
	R3    T3
	R4    T4
	R5    T5
	R6    T6
	R7    T7
	R8    T8
	R9    T9
	R10   T10
	R11   T11
	R12   T12
	R13   T13
	R14   T14
	R15   T15
	R16   T16
	Error error
}

func (rtp ResultTuple16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16]) Extract() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.R12, rtp.R13, rtp.R14, rtp.R15, rtp.R16, rtp.Error
}

func (rtp ResultTuple16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16]) Ensure() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16) {
	return rtp.ensure(2)
}

func (rtp ResultTuple16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16]) ensure(skip int) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.R12, rtp.R13, rtp.R14, rtp.R15, rtp.R16
}

func Result16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16 any](future async.Future) (rtp ResultTuple16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16]) {
	return Parse16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16](<-future.Chan())
}
