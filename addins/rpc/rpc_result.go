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
	"context"

	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/exception"
)

// ResultValues 保存未指定返回值类型的 RPC 结果及错误。
type ResultValues struct {
	Values []any
	Error  error
}

// Extract 返回结果值和错误。
func (rvs ResultValues) Extract() ([]any, error) {
	return rvs.Values, rvs.Error
}

// Ensure 返回结果值；结果包含错误时 panic。
func (rvs ResultValues) Ensure() []any {
	return rvs.ensure(2)
}

func (rvs ResultValues) ensure(skip int) []any {
	if rvs.Error != nil {
		exception.PanicSkip(skip, rvs.Error)
	}
	return rvs.Values
}

// Results 等待 Future 的首个结果，并将其解析为未指定类型的返回值列表。
func Results(future async.Future) (rvs ResultValues) {
	return ParseResults(future.Wait(context.Background()))
}

// ResultTupleVoid 保存无返回值 RPC 的错误。
type ResultTupleVoid struct {
	Error error
}

// Extract 返回 RPC 错误。
func (rtp ResultTupleVoid) Extract() error {
	return rtp.Error
}

// Ensure 在结果包含错误时 panic。
func (rtp ResultTupleVoid) Ensure() {
	rtp.ensure(2)
}

func (rtp ResultTupleVoid) ensure(skip int) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
}

// ResultVoid 等待 Future 的首个结果，并将其解析为无返回值结果。
func ResultVoid(future async.Future) (rtp ResultTupleVoid) {
	return ParseVoid(future.Wait(context.Background()))
}

// ResultTuple1 保存一个指定类型的 RPC 返回值及错误。
type ResultTuple1[T1 any] struct {
	R1    T1
	Error error
}

// Extract 返回一个结果值和错误。
func (rtp ResultTuple1[T1]) Extract() (T1, error) {
	return rtp.R1, rtp.Error
}

// Ensure 返回一个结果值；结果包含错误时 panic。
func (rtp ResultTuple1[T1]) Ensure() T1 {
	return rtp.ensure(2)
}

func (rtp ResultTuple1[T1]) ensure(skip int) T1 {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1
}

// Result1 等待 Future 的首个结果，并解析一个指定类型的返回值。
func Result1[T1 any](future async.Future) (rtp ResultTuple1[T1]) {
	return Parse1[T1](future.Wait(context.Background()))
}

// ResultTuple2 保存两个指定类型的 RPC 返回值及错误。
type ResultTuple2[T1, T2 any] struct {
	R1    T1
	R2    T2
	Error error
}

// Extract 返回两个结果值和错误。
func (rtp ResultTuple2[T1, T2]) Extract() (T1, T2, error) {
	return rtp.R1, rtp.R2, rtp.Error
}

// Ensure 返回两个结果值；结果包含错误时 panic。
func (rtp ResultTuple2[T1, T2]) Ensure() (T1, T2) {
	return rtp.ensure(2)
}

func (rtp ResultTuple2[T1, T2]) ensure(skip int) (T1, T2) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2
}

// Result2 等待 Future 的首个结果，并解析两个指定类型的返回值。
func Result2[T1, T2 any](future async.Future) (rtp ResultTuple2[T1, T2]) {
	return Parse2[T1, T2](future.Wait(context.Background()))
}

// ResultTuple3 保存三个指定类型的 RPC 返回值及错误。
type ResultTuple3[T1, T2, T3 any] struct {
	R1    T1
	R2    T2
	R3    T3
	Error error
}

// Extract 返回三个结果值和错误。
func (rtp ResultTuple3[T1, T2, T3]) Extract() (T1, T2, T3, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.Error
}

// Ensure 返回三个结果值；结果包含错误时 panic。
func (rtp ResultTuple3[T1, T2, T3]) Ensure() (T1, T2, T3) {
	return rtp.ensure(2)
}

func (rtp ResultTuple3[T1, T2, T3]) ensure(skip int) (T1, T2, T3) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3
}

// Result3 等待 Future 的首个结果，并解析三个指定类型的返回值。
func Result3[T1, T2, T3 any](future async.Future) (rtp ResultTuple3[T1, T2, T3]) {
	return Parse3[T1, T2, T3](future.Wait(context.Background()))
}

// ResultTuple4 保存四个指定类型的 RPC 返回值及错误。
type ResultTuple4[T1, T2, T3, T4 any] struct {
	R1    T1
	R2    T2
	R3    T3
	R4    T4
	Error error
}

// Extract 返回四个结果值和错误。
func (rtp ResultTuple4[T1, T2, T3, T4]) Extract() (T1, T2, T3, T4, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.Error
}

// Ensure 返回四个结果值；结果包含错误时 panic。
func (rtp ResultTuple4[T1, T2, T3, T4]) Ensure() (T1, T2, T3, T4) {
	return rtp.ensure(2)
}

func (rtp ResultTuple4[T1, T2, T3, T4]) ensure(skip int) (T1, T2, T3, T4) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4
}

// Result4 等待 Future 的首个结果，并解析四个指定类型的返回值。
func Result4[T1, T2, T3, T4 any](future async.Future) (rtp ResultTuple4[T1, T2, T3, T4]) {
	return Parse4[T1, T2, T3, T4](future.Wait(context.Background()))
}

// ResultTuple5 保存五个指定类型的 RPC 返回值及错误。
type ResultTuple5[T1, T2, T3, T4, T5 any] struct {
	R1    T1
	R2    T2
	R3    T3
	R4    T4
	R5    T5
	Error error
}

// Extract 返回五个结果值和错误。
func (rtp ResultTuple5[T1, T2, T3, T4, T5]) Extract() (T1, T2, T3, T4, T5, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.Error
}

// Ensure 返回五个结果值；结果包含错误时 panic。
func (rtp ResultTuple5[T1, T2, T3, T4, T5]) Ensure() (T1, T2, T3, T4, T5) {
	return rtp.ensure(2)
}

func (rtp ResultTuple5[T1, T2, T3, T4, T5]) ensure(skip int) (T1, T2, T3, T4, T5) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5
}

// Result5 等待 Future 的首个结果，并解析五个指定类型的返回值。
func Result5[T1, T2, T3, T4, T5 any](future async.Future) (rtp ResultTuple5[T1, T2, T3, T4, T5]) {
	return Parse5[T1, T2, T3, T4, T5](future.Wait(context.Background()))
}

// ResultTuple6 保存六个指定类型的 RPC 返回值及错误。
type ResultTuple6[T1, T2, T3, T4, T5, T6 any] struct {
	R1    T1
	R2    T2
	R3    T3
	R4    T4
	R5    T5
	R6    T6
	Error error
}

// Extract 返回六个结果值和错误。
func (rtp ResultTuple6[T1, T2, T3, T4, T5, T6]) Extract() (T1, T2, T3, T4, T5, T6, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.Error
}

// Ensure 返回六个结果值；结果包含错误时 panic。
func (rtp ResultTuple6[T1, T2, T3, T4, T5, T6]) Ensure() (T1, T2, T3, T4, T5, T6) {
	return rtp.ensure(2)
}

func (rtp ResultTuple6[T1, T2, T3, T4, T5, T6]) ensure(skip int) (T1, T2, T3, T4, T5, T6) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6
}

// Result6 等待 Future 的首个结果，并解析六个指定类型的返回值。
func Result6[T1, T2, T3, T4, T5, T6 any](future async.Future) (rtp ResultTuple6[T1, T2, T3, T4, T5, T6]) {
	return Parse6[T1, T2, T3, T4, T5, T6](future.Wait(context.Background()))
}

// ResultTuple7 保存七个指定类型的 RPC 返回值及错误。
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

// Extract 返回七个结果值和错误。
func (rtp ResultTuple7[T1, T2, T3, T4, T5, T6, T7]) Extract() (T1, T2, T3, T4, T5, T6, T7, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.Error
}

// Ensure 返回七个结果值；结果包含错误时 panic。
func (rtp ResultTuple7[T1, T2, T3, T4, T5, T6, T7]) Ensure() (T1, T2, T3, T4, T5, T6, T7) {
	return rtp.ensure(2)
}

func (rtp ResultTuple7[T1, T2, T3, T4, T5, T6, T7]) ensure(skip int) (T1, T2, T3, T4, T5, T6, T7) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7
}

// Result7 等待 Future 的首个结果，并解析七个指定类型的返回值。
func Result7[T1, T2, T3, T4, T5, T6, T7 any](future async.Future) (rtp ResultTuple7[T1, T2, T3, T4, T5, T6, T7]) {
	return Parse7[T1, T2, T3, T4, T5, T6, T7](future.Wait(context.Background()))
}

// ResultTuple8 保存八个指定类型的 RPC 返回值及错误。
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

// Extract 返回八个结果值和错误。
func (rtp ResultTuple8[T1, T2, T3, T4, T5, T6, T7, T8]) Extract() (T1, T2, T3, T4, T5, T6, T7, T8, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.Error
}

// Ensure 返回八个结果值；结果包含错误时 panic。
func (rtp ResultTuple8[T1, T2, T3, T4, T5, T6, T7, T8]) Ensure() (T1, T2, T3, T4, T5, T6, T7, T8) {
	return rtp.ensure(2)
}

func (rtp ResultTuple8[T1, T2, T3, T4, T5, T6, T7, T8]) ensure(skip int) (T1, T2, T3, T4, T5, T6, T7, T8) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8
}

// Result8 等待 Future 的首个结果，并解析八个指定类型的返回值。
func Result8[T1, T2, T3, T4, T5, T6, T7, T8 any](future async.Future) (rtp ResultTuple8[T1, T2, T3, T4, T5, T6, T7, T8]) {
	return Parse8[T1, T2, T3, T4, T5, T6, T7, T8](future.Wait(context.Background()))
}

// ResultTuple9 保存九个指定类型的 RPC 返回值及错误。
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

// Extract 返回九个结果值和错误。
func (rtp ResultTuple9[T1, T2, T3, T4, T5, T6, T7, T8, T9]) Extract() (T1, T2, T3, T4, T5, T6, T7, T8, T9, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.Error
}

// Ensure 返回九个结果值；结果包含错误时 panic。
func (rtp ResultTuple9[T1, T2, T3, T4, T5, T6, T7, T8, T9]) Ensure() (T1, T2, T3, T4, T5, T6, T7, T8, T9) {
	return rtp.ensure(2)
}

func (rtp ResultTuple9[T1, T2, T3, T4, T5, T6, T7, T8, T9]) ensure(skip int) (T1, T2, T3, T4, T5, T6, T7, T8, T9) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9
}

// Result9 等待 Future 的首个结果，并解析九个指定类型的返回值。
func Result9[T1, T2, T3, T4, T5, T6, T7, T8, T9 any](future async.Future) (rtp ResultTuple9[T1, T2, T3, T4, T5, T6, T7, T8, T9]) {
	return Parse9[T1, T2, T3, T4, T5, T6, T7, T8, T9](future.Wait(context.Background()))
}

// ResultTuple10 保存十个指定类型的 RPC 返回值及错误。
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

// Extract 返回十个结果值和错误。
func (rtp ResultTuple10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) Extract() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.Error
}

// Ensure 返回十个结果值；结果包含错误时 panic。
func (rtp ResultTuple10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) Ensure() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10) {
	return rtp.ensure(2)
}

func (rtp ResultTuple10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) ensure(skip int) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10
}

// Result10 等待 Future 的首个结果，并解析十个指定类型的返回值。
func Result10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10 any](future async.Future) (rtp ResultTuple10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) {
	return Parse10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10](future.Wait(context.Background()))
}

// ResultTuple11 保存十一个指定类型的 RPC 返回值及错误。
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

// Extract 返回十一个结果值和错误。
func (rtp ResultTuple11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]) Extract() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.Error
}

// Ensure 返回十一个结果值；结果包含错误时 panic。
func (rtp ResultTuple11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]) Ensure() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11) {
	return rtp.ensure(2)
}

func (rtp ResultTuple11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]) ensure(skip int) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11
}

// Result11 等待 Future 的首个结果，并解析十一个指定类型的返回值。
func Result11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11 any](future async.Future) (rtp ResultTuple11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]) {
	return Parse11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11](future.Wait(context.Background()))
}

// ResultTuple12 保存十二个指定类型的 RPC 返回值及错误。
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

// Extract 返回十二个结果值和错误。
func (rtp ResultTuple12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) Extract() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.R12, rtp.Error
}

// Ensure 返回十二个结果值；结果包含错误时 panic。
func (rtp ResultTuple12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) Ensure() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12) {
	return rtp.ensure(2)
}

func (rtp ResultTuple12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) ensure(skip int) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.R12
}

// Result12 等待 Future 的首个结果，并解析十二个指定类型的返回值。
func Result12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12 any](future async.Future) (rtp ResultTuple12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) {
	return Parse12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12](future.Wait(context.Background()))
}

// ResultTuple13 保存十三个指定类型的 RPC 返回值及错误。
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

// Extract 返回十三个结果值和错误。
func (rtp ResultTuple13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) Extract() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.R12, rtp.R13, rtp.Error
}

// Ensure 返回十三个结果值；结果包含错误时 panic。
func (rtp ResultTuple13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) Ensure() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13) {
	return rtp.ensure(2)
}

func (rtp ResultTuple13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) ensure(skip int) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.R12, rtp.R13
}

// Result13 等待 Future 的首个结果，并解析十三个指定类型的返回值。
func Result13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13 any](future async.Future) (rtp ResultTuple13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) {
	return Parse13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13](future.Wait(context.Background()))
}

// ResultTuple14 保存十四个指定类型的 RPC 返回值及错误。
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

// Extract 返回十四个结果值和错误。
func (rtp ResultTuple14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) Extract() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.R12, rtp.R13, rtp.R14, rtp.Error
}

// Ensure 返回十四个结果值；结果包含错误时 panic。
func (rtp ResultTuple14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) Ensure() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14) {
	return rtp.ensure(2)
}

func (rtp ResultTuple14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) ensure(skip int) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.R12, rtp.R13, rtp.R14
}

// Result14 等待 Future 的首个结果，并解析十四个指定类型的返回值。
func Result14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14 any](future async.Future) (rtp ResultTuple14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) {
	return Parse14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14](future.Wait(context.Background()))
}

// ResultTuple15 保存十五个指定类型的 RPC 返回值及错误。
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

// Extract 返回十五个结果值和错误。
func (rtp ResultTuple15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) Extract() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.R12, rtp.R13, rtp.R14, rtp.R15, rtp.Error
}

// Ensure 返回十五个结果值；结果包含错误时 panic。
func (rtp ResultTuple15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) Ensure() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15) {
	return rtp.ensure(2)
}

func (rtp ResultTuple15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) ensure(skip int) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.R12, rtp.R13, rtp.R14, rtp.R15
}

// Result15 等待 Future 的首个结果，并解析十五个指定类型的返回值。
func Result15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15 any](future async.Future) (rtp ResultTuple15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) {
	return Parse15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15](future.Wait(context.Background()))
}

// ResultTuple16 保存十六个指定类型的 RPC 返回值及错误。
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

// Extract 返回十六个结果值和错误。
func (rtp ResultTuple16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16]) Extract() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16, error) {
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.R12, rtp.R13, rtp.R14, rtp.R15, rtp.R16, rtp.Error
}

// Ensure 返回十六个结果值；结果包含错误时 panic。
func (rtp ResultTuple16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16]) Ensure() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16) {
	return rtp.ensure(2)
}

func (rtp ResultTuple16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16]) ensure(skip int) (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16) {
	if rtp.Error != nil {
		exception.PanicSkip(skip, rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.R12, rtp.R13, rtp.R14, rtp.R15, rtp.R16
}

// Result16 等待 Future 的首个结果，并解析十六个指定类型的返回值。
func Result16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16 any](future async.Future) (rtp ResultTuple16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16]) {
	return Parse16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16](future.Wait(context.Background()))
}
