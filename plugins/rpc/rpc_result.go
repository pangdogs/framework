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
	"errors"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/types"
	"git.golaxy.org/framework/net/gap/variant"
	"reflect"
)

var (
	ErrMethodResultCountMismatch = errors.New("rpc: method result count mismatch")
	ErrMethodResultTypeMismatch  = errors.New("rpc: method result type mismatch")
)

func parseRet[T any](retArr variant.Array, idx int) (T, error) {
	v := retArr[idx]

	ret, ok := v.Value.Indirect().(T)
	if ok {
		return ret, nil
	}

	retRV, err := v.Convert(reflect.TypeFor[T]())
	if err != nil {
		return types.ZeroT[T](), ErrMethodResultTypeMismatch
	}

	if retRV.IsNil() {
		return types.ZeroT[T](), nil
	}

	return retRV.Interface().(T), nil
}

type ResultValues struct {
	Values []any
	Error  error
}

func (rvs ResultValues) Extract() ([]any, error) {
	return rvs.Values, rvs.Error
}

func (rvs ResultValues) Ensure() []any {
	if rvs.Error != nil {
		panic(rvs.Error)
	}
	return rvs.Values
}

func Results(ret async.Ret) (rvs ResultValues) {
	if !ret.OK() {
		rvs.Error = ret.Error
		return
	}

	if ret.Value == nil {
		return
	}

	retArr, ok := ret.Value.(variant.Array)
	if !ok || len(retArr) < 1 {
		return
	}

	rets := make([]any, len(retArr))

	for i := range rets {
		rets[i] = retArr[i].Value.Indirect()
	}

	rvs.Values = rets
	return
}

type ResultTuple0 struct {
	Error error
}

func (rtp ResultTuple0) Extract() error {
	return rtp.Error
}

func (rtp ResultTuple0) Ensure() {
	if rtp.Error != nil {
		panic(rtp.Error)
	}
}

func ResultVoid(ret async.Ret) (rtp ResultTuple0) {
	if !ret.OK() {
		rtp.Error = ret.Error
		return
	}
	return
}

type ResultTuple1[T1 any] struct {
	R1    T1
	Error error
}

func (rtp ResultTuple1[T1]) Extract() (T1, error) {
	return rtp.R1, rtp.Error
}

func (rtp ResultTuple1[T1]) Ensure() T1 {
	if rtp.Error != nil {
		panic(rtp.Error)
	}
	return rtp.R1
}

func Result1[T1 any](ret async.Ret) (rtp ResultTuple1[T1]) {
	if !ret.OK() {
		rtp.Error = ret.Error
		return
	}

	if ret.Value == nil {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	retArr, ok := ret.Value.(variant.Array)
	if !ok || len(retArr) < 1 {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	r1, err := parseRet[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	rtp.R1 = r1
	return
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
	if rtp.Error != nil {
		panic(rtp.Error)
	}
	return rtp.R1, rtp.R2
}

func Result2[T1, T2 any](ret async.Ret) (rtp ResultTuple2[T1, T2]) {
	if !ret.OK() {
		rtp.Error = ret.Error
		return
	}

	if ret.Value == nil {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	retArr, ok := ret.Value.(variant.Array)
	if !ok || len(retArr) < 2 {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	r1, err := parseRet[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseRet[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	rtp.R1 = r1
	rtp.R2 = r2
	return
}

func AssertRet2[T1, T2 any](ret async.Ret) (T1, T2) {
	return Result2[T1, T2](ret).Ensure()
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
	if rtp.Error != nil {
		panic(rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3
}

func Result3[T1, T2, T3 any](ret async.Ret) (rtp ResultTuple3[T1, T2, T3]) {
	if !ret.OK() {
		rtp.Error = ret.Error
		return
	}

	if ret.Value == nil {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	retArr, ok := ret.Value.(variant.Array)
	if !ok || len(retArr) < 3 {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	r1, err := parseRet[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseRet[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseRet[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	rtp.R1 = r1
	rtp.R2 = r2
	rtp.R3 = r3
	return
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
	if rtp.Error != nil {
		panic(rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4
}

func Result4[T1, T2, T3, T4 any](ret async.Ret) (rtp ResultTuple4[T1, T2, T3, T4]) {
	if !ret.OK() {
		rtp.Error = ret.Error
		return
	}

	if ret.Value == nil {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	retArr, ok := ret.Value.(variant.Array)
	if !ok || len(retArr) < 4 {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	r1, err := parseRet[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseRet[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseRet[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	r4, err := parseRet[T4](retArr, 3)
	if err != nil {
		rtp.Error = err
		return
	}

	rtp.R1 = r1
	rtp.R2 = r2
	rtp.R3 = r3
	rtp.R4 = r4
	return
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
	if rtp.Error != nil {
		panic(rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5
}

func Result5[T1, T2, T3, T4, T5 any](ret async.Ret) (rtp ResultTuple5[T1, T2, T3, T4, T5]) {
	if !ret.OK() {
		rtp.Error = ret.Error
		return
	}

	if ret.Value == nil {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	retArr, ok := ret.Value.(variant.Array)
	if !ok || len(retArr) < 5 {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	r1, err := parseRet[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseRet[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseRet[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	r4, err := parseRet[T4](retArr, 3)
	if err != nil {
		rtp.Error = err
		return
	}

	r5, err := parseRet[T5](retArr, 4)
	if err != nil {
		rtp.Error = err
		return
	}

	rtp.R1 = r1
	rtp.R2 = r2
	rtp.R3 = r3
	rtp.R4 = r4
	rtp.R5 = r5
	return
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
	if rtp.Error != nil {
		panic(rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6
}

func Result6[T1, T2, T3, T4, T5, T6 any](ret async.Ret) (rtp ResultTuple6[T1, T2, T3, T4, T5, T6]) {
	if !ret.OK() {
		rtp.Error = ret.Error
		return
	}

	if ret.Value == nil {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	retArr, ok := ret.Value.(variant.Array)
	if !ok || len(retArr) < 6 {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	r1, err := parseRet[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseRet[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseRet[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	r4, err := parseRet[T4](retArr, 3)
	if err != nil {
		rtp.Error = err
		return
	}

	r5, err := parseRet[T5](retArr, 4)
	if err != nil {
		rtp.Error = err
		return
	}

	r6, err := parseRet[T6](retArr, 5)
	if err != nil {
		rtp.Error = err
		return
	}

	rtp.R1 = r1
	rtp.R2 = r2
	rtp.R3 = r3
	rtp.R4 = r4
	rtp.R5 = r5
	rtp.R6 = r6
	return
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
	if rtp.Error != nil {
		panic(rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7
}

func Result7[T1, T2, T3, T4, T5, T6, T7 any](ret async.Ret) (rtp ResultTuple7[T1, T2, T3, T4, T5, T6, T7]) {
	if !ret.OK() {
		rtp.Error = ret.Error
		return
	}

	if ret.Value == nil {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	retArr, ok := ret.Value.(variant.Array)
	if !ok || len(retArr) < 7 {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	r1, err := parseRet[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseRet[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseRet[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	r4, err := parseRet[T4](retArr, 3)
	if err != nil {
		rtp.Error = err
		return
	}

	r5, err := parseRet[T5](retArr, 4)
	if err != nil {
		rtp.Error = err
		return
	}

	r6, err := parseRet[T6](retArr, 5)
	if err != nil {
		rtp.Error = err
		return
	}

	r7, err := parseRet[T7](retArr, 6)
	if err != nil {
		rtp.Error = err
		return
	}

	rtp.R1 = r1
	rtp.R2 = r2
	rtp.R3 = r3
	rtp.R4 = r4
	rtp.R5 = r5
	rtp.R6 = r6
	rtp.R7 = r7
	return
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
	if rtp.Error != nil {
		panic(rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8
}

func Result8[T1, T2, T3, T4, T5, T6, T7, T8 any](ret async.Ret) (rtp ResultTuple8[T1, T2, T3, T4, T5, T6, T7, T8]) {
	if !ret.OK() {
		rtp.Error = ret.Error
		return
	}

	if ret.Value == nil {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	retArr, ok := ret.Value.(variant.Array)
	if !ok || len(retArr) < 8 {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	r1, err := parseRet[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseRet[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseRet[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	r4, err := parseRet[T4](retArr, 3)
	if err != nil {
		rtp.Error = err
		return
	}

	r5, err := parseRet[T5](retArr, 4)
	if err != nil {
		rtp.Error = err
		return
	}

	r6, err := parseRet[T6](retArr, 5)
	if err != nil {
		rtp.Error = err
		return
	}

	r7, err := parseRet[T7](retArr, 6)
	if err != nil {
		rtp.Error = err
		return
	}

	r8, err := parseRet[T8](retArr, 7)
	if err != nil {
		rtp.Error = err
		return
	}

	rtp.R1 = r1
	rtp.R2 = r2
	rtp.R3 = r3
	rtp.R4 = r4
	rtp.R5 = r5
	rtp.R6 = r6
	rtp.R7 = r7
	rtp.R8 = r8
	return
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
	if rtp.Error != nil {
		panic(rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9
}

func Result9[T1, T2, T3, T4, T5, T6, T7, T8, T9 any](ret async.Ret) (rtp ResultTuple9[T1, T2, T3, T4, T5, T6, T7, T8, T9]) {
	if !ret.OK() {
		rtp.Error = ret.Error
		return
	}

	if ret.Value == nil {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	retArr, ok := ret.Value.(variant.Array)
	if !ok || len(retArr) < 9 {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	r1, err := parseRet[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseRet[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseRet[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	r4, err := parseRet[T4](retArr, 3)
	if err != nil {
		rtp.Error = err
		return
	}

	r5, err := parseRet[T5](retArr, 4)
	if err != nil {
		rtp.Error = err
		return
	}

	r6, err := parseRet[T6](retArr, 5)
	if err != nil {
		rtp.Error = err
		return
	}

	r7, err := parseRet[T7](retArr, 6)
	if err != nil {
		rtp.Error = err
		return
	}

	r8, err := parseRet[T8](retArr, 7)
	if err != nil {
		rtp.Error = err
		return
	}

	r9, err := parseRet[T9](retArr, 8)
	if err != nil {
		rtp.Error = err
		return
	}

	rtp.R1 = r1
	rtp.R2 = r2
	rtp.R3 = r3
	rtp.R4 = r4
	rtp.R5 = r5
	rtp.R6 = r6
	rtp.R7 = r7
	rtp.R8 = r8
	rtp.R9 = r9
	return
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
	if rtp.Error != nil {
		panic(rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10
}

func Result10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10 any](ret async.Ret) (rtp ResultTuple10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) {
	if !ret.OK() {
		rtp.Error = ret.Error
		return
	}

	if ret.Value == nil {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	retArr, ok := ret.Value.(variant.Array)
	if !ok || len(retArr) < 10 {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	r1, err := parseRet[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseRet[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseRet[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	r4, err := parseRet[T4](retArr, 3)
	if err != nil {
		rtp.Error = err
		return
	}

	r5, err := parseRet[T5](retArr, 4)
	if err != nil {
		rtp.Error = err
		return
	}

	r6, err := parseRet[T6](retArr, 5)
	if err != nil {
		rtp.Error = err
		return
	}

	r7, err := parseRet[T7](retArr, 6)
	if err != nil {
		rtp.Error = err
		return
	}

	r8, err := parseRet[T8](retArr, 7)
	if err != nil {
		rtp.Error = err
		return
	}

	r9, err := parseRet[T9](retArr, 8)
	if err != nil {
		rtp.Error = err
		return
	}

	r10, err := parseRet[T10](retArr, 9)
	if err != nil {
		rtp.Error = err
		return
	}

	rtp.R1 = r1
	rtp.R2 = r2
	rtp.R3 = r3
	rtp.R4 = r4
	rtp.R5 = r5
	rtp.R6 = r6
	rtp.R7 = r7
	rtp.R8 = r8
	rtp.R9 = r9
	rtp.R10 = r10
	return
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
	if rtp.Error != nil {
		panic(rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11
}

func Result11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11 any](ret async.Ret) (rtp ResultTuple11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]) {
	if !ret.OK() {
		rtp.Error = ret.Error
		return
	}

	if ret.Value == nil {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	retArr, ok := ret.Value.(variant.Array)
	if !ok || len(retArr) < 11 {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	r1, err := parseRet[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseRet[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseRet[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	r4, err := parseRet[T4](retArr, 3)
	if err != nil {
		rtp.Error = err
		return
	}

	r5, err := parseRet[T5](retArr, 4)
	if err != nil {
		rtp.Error = err
		return
	}

	r6, err := parseRet[T6](retArr, 5)
	if err != nil {
		rtp.Error = err
		return
	}

	r7, err := parseRet[T7](retArr, 6)
	if err != nil {
		rtp.Error = err
		return
	}

	r8, err := parseRet[T8](retArr, 7)
	if err != nil {
		rtp.Error = err
		return
	}

	r9, err := parseRet[T9](retArr, 8)
	if err != nil {
		rtp.Error = err
		return
	}

	r10, err := parseRet[T10](retArr, 9)
	if err != nil {
		rtp.Error = err
		return
	}

	r11, err := parseRet[T11](retArr, 10)
	if err != nil {
		rtp.Error = err
		return
	}

	rtp.R1 = r1
	rtp.R2 = r2
	rtp.R3 = r3
	rtp.R4 = r4
	rtp.R5 = r5
	rtp.R6 = r6
	rtp.R7 = r7
	rtp.R8 = r8
	rtp.R9 = r9
	rtp.R10 = r10
	rtp.R11 = r11
	return
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
	if rtp.Error != nil {
		panic(rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.R12
}

func Result12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12 any](ret async.Ret) (rtp ResultTuple12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) {
	if !ret.OK() {
		rtp.Error = ret.Error
		return
	}

	retArr, ok := ret.Value.(variant.Array)
	if !ok || len(retArr) < 12 {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	r1, err := parseRet[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseRet[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseRet[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	r4, err := parseRet[T4](retArr, 3)
	if err != nil {
		rtp.Error = err
		return
	}

	r5, err := parseRet[T5](retArr, 4)
	if err != nil {
		rtp.Error = err
		return
	}

	r6, err := parseRet[T6](retArr, 5)
	if err != nil {
		rtp.Error = err
		return
	}

	r7, err := parseRet[T7](retArr, 6)
	if err != nil {
		rtp.Error = err
		return
	}

	r8, err := parseRet[T8](retArr, 7)
	if err != nil {
		rtp.Error = err
		return
	}

	r9, err := parseRet[T9](retArr, 8)
	if err != nil {
		rtp.Error = err
		return
	}

	r10, err := parseRet[T10](retArr, 9)
	if err != nil {
		rtp.Error = err
		return
	}

	r11, err := parseRet[T11](retArr, 10)
	if err != nil {
		rtp.Error = err
		return
	}

	r12, err := parseRet[T12](retArr, 11)
	if err != nil {
		rtp.Error = err
		return
	}

	rtp.R1 = r1
	rtp.R2 = r2
	rtp.R3 = r3
	rtp.R4 = r4
	rtp.R5 = r5
	rtp.R6 = r6
	rtp.R7 = r7
	rtp.R8 = r8
	rtp.R9 = r9
	rtp.R10 = r10
	rtp.R11 = r11
	rtp.R12 = r12
	return
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
	if rtp.Error != nil {
		panic(rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.R12, rtp.R13
}

func Result13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13 any](ret async.Ret) (rtp ResultTuple13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) {
	if !ret.OK() {
		rtp.Error = ret.Error
		return
	}

	retArr, ok := ret.Value.(variant.Array)
	if !ok || len(retArr) < 13 {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	r1, err := parseRet[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseRet[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseRet[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	r4, err := parseRet[T4](retArr, 3)
	if err != nil {
		rtp.Error = err
		return
	}

	r5, err := parseRet[T5](retArr, 4)
	if err != nil {
		rtp.Error = err
		return
	}

	r6, err := parseRet[T6](retArr, 5)
	if err != nil {
		rtp.Error = err
		return
	}

	r7, err := parseRet[T7](retArr, 6)
	if err != nil {
		rtp.Error = err
		return
	}

	r8, err := parseRet[T8](retArr, 7)
	if err != nil {
		rtp.Error = err
		return
	}

	r9, err := parseRet[T9](retArr, 8)
	if err != nil {
		rtp.Error = err
		return
	}

	r10, err := parseRet[T10](retArr, 9)
	if err != nil {
		rtp.Error = err
		return
	}

	r11, err := parseRet[T11](retArr, 10)
	if err != nil {
		rtp.Error = err
		return
	}

	r12, err := parseRet[T12](retArr, 11)
	if err != nil {
		rtp.Error = err
		return
	}

	r13, err := parseRet[T13](retArr, 12)
	if err != nil {
		rtp.Error = err
		return
	}

	rtp.R1 = r1
	rtp.R2 = r2
	rtp.R3 = r3
	rtp.R4 = r4
	rtp.R5 = r5
	rtp.R6 = r6
	rtp.R7 = r7
	rtp.R8 = r8
	rtp.R9 = r9
	rtp.R10 = r10
	rtp.R11 = r11
	rtp.R12 = r12
	rtp.R13 = r13
	return
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
	if rtp.Error != nil {
		panic(rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.R12, rtp.R13, rtp.R14
}

func Result14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14 any](ret async.Ret) (rtp ResultTuple14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) {
	if !ret.OK() {
		rtp.Error = ret.Error
		return
	}

	retArr, ok := ret.Value.(variant.Array)
	if !ok || len(retArr) < 14 {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	r1, err := parseRet[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseRet[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseRet[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	r4, err := parseRet[T4](retArr, 3)
	if err != nil {
		rtp.Error = err
		return
	}

	r5, err := parseRet[T5](retArr, 4)
	if err != nil {
		rtp.Error = err
		return
	}

	r6, err := parseRet[T6](retArr, 5)
	if err != nil {
		rtp.Error = err
		return
	}

	r7, err := parseRet[T7](retArr, 6)
	if err != nil {
		rtp.Error = err
		return
	}

	r8, err := parseRet[T8](retArr, 7)
	if err != nil {
		rtp.Error = err
		return
	}

	r9, err := parseRet[T9](retArr, 8)
	if err != nil {
		rtp.Error = err
		return
	}

	r10, err := parseRet[T10](retArr, 9)
	if err != nil {
		rtp.Error = err
		return
	}

	r11, err := parseRet[T11](retArr, 10)
	if err != nil {
		rtp.Error = err
		return
	}

	r12, err := parseRet[T12](retArr, 11)
	if err != nil {
		rtp.Error = err
		return
	}

	r13, err := parseRet[T13](retArr, 12)
	if err != nil {
		rtp.Error = err
		return
	}

	r14, err := parseRet[T14](retArr, 13)
	if err != nil {
		rtp.Error = err
		return
	}

	rtp.R1 = r1
	rtp.R2 = r2
	rtp.R3 = r3
	rtp.R4 = r4
	rtp.R5 = r5
	rtp.R6 = r6
	rtp.R7 = r7
	rtp.R8 = r8
	rtp.R9 = r9
	rtp.R10 = r10
	rtp.R11 = r11
	rtp.R12 = r12
	rtp.R13 = r13
	rtp.R14 = r14
	return
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
	if rtp.Error != nil {
		panic(rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.R12, rtp.R13, rtp.R14, rtp.R15
}

func Result15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15 any](ret async.Ret) (rtp ResultTuple15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) {
	if !ret.OK() {
		rtp.Error = ret.Error
		return
	}

	retArr, ok := ret.Value.(variant.Array)
	if !ok || len(retArr) < 15 {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	r1, err := parseRet[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseRet[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseRet[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	r4, err := parseRet[T4](retArr, 3)
	if err != nil {
		rtp.Error = err
		return
	}

	r5, err := parseRet[T5](retArr, 4)
	if err != nil {
		rtp.Error = err
		return
	}

	r6, err := parseRet[T6](retArr, 5)
	if err != nil {
		rtp.Error = err
		return
	}

	r7, err := parseRet[T7](retArr, 6)
	if err != nil {
		rtp.Error = err
		return
	}

	r8, err := parseRet[T8](retArr, 7)
	if err != nil {
		rtp.Error = err
		return
	}

	r9, err := parseRet[T9](retArr, 8)
	if err != nil {
		rtp.Error = err
		return
	}

	r10, err := parseRet[T10](retArr, 9)
	if err != nil {
		rtp.Error = err
		return
	}

	r11, err := parseRet[T11](retArr, 10)
	if err != nil {
		rtp.Error = err
		return
	}

	r12, err := parseRet[T12](retArr, 11)
	if err != nil {
		rtp.Error = err
		return
	}

	r13, err := parseRet[T13](retArr, 12)
	if err != nil {
		rtp.Error = err
		return
	}

	r14, err := parseRet[T14](retArr, 13)
	if err != nil {
		rtp.Error = err
		return
	}

	r15, err := parseRet[T15](retArr, 14)
	if err != nil {
		rtp.Error = err
		return
	}

	rtp.R1 = r1
	rtp.R2 = r2
	rtp.R3 = r3
	rtp.R4 = r4
	rtp.R5 = r5
	rtp.R6 = r6
	rtp.R7 = r7
	rtp.R8 = r8
	rtp.R9 = r9
	rtp.R10 = r10
	rtp.R11 = r11
	rtp.R12 = r12
	rtp.R13 = r13
	rtp.R14 = r14
	rtp.R15 = r15
	return
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
	if rtp.Error != nil {
		panic(rtp.Error)
	}
	return rtp.R1, rtp.R2, rtp.R3, rtp.R4, rtp.R5, rtp.R6, rtp.R7, rtp.R8, rtp.R9, rtp.R10, rtp.R11, rtp.R12, rtp.R13, rtp.R14, rtp.R15, rtp.R16
}

func Result16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16 any](ret async.Ret) (rtp ResultTuple16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16]) {
	if !ret.OK() {
		rtp.Error = ret.Error
		return
	}

	retArr, ok := ret.Value.(variant.Array)
	if !ok || len(retArr) < 16 {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	r1, err := parseRet[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseRet[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseRet[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	r4, err := parseRet[T4](retArr, 3)
	if err != nil {
		rtp.Error = err
		return
	}

	r5, err := parseRet[T5](retArr, 4)
	if err != nil {
		rtp.Error = err
		return
	}

	r6, err := parseRet[T6](retArr, 5)
	if err != nil {
		rtp.Error = err
		return
	}

	r7, err := parseRet[T7](retArr, 6)
	if err != nil {
		rtp.Error = err
		return
	}

	r8, err := parseRet[T8](retArr, 7)
	if err != nil {
		rtp.Error = err
		return
	}

	r9, err := parseRet[T9](retArr, 8)
	if err != nil {
		rtp.Error = err
		return
	}

	r10, err := parseRet[T10](retArr, 9)
	if err != nil {
		rtp.Error = err
		return
	}

	r11, err := parseRet[T11](retArr, 10)
	if err != nil {
		rtp.Error = err
		return
	}

	r12, err := parseRet[T12](retArr, 11)
	if err != nil {
		rtp.Error = err
		return
	}

	r13, err := parseRet[T13](retArr, 12)
	if err != nil {
		rtp.Error = err
		return
	}

	r14, err := parseRet[T14](retArr, 13)
	if err != nil {
		rtp.Error = err
		return
	}

	r15, err := parseRet[T15](retArr, 14)
	if err != nil {
		rtp.Error = err
		return
	}

	r16, err := parseRet[T16](retArr, 15)
	if err != nil {
		rtp.Error = err
		return
	}

	rtp.R1 = r1
	rtp.R2 = r2
	rtp.R3 = r3
	rtp.R4 = r4
	rtp.R5 = r5
	rtp.R6 = r6
	rtp.R7 = r7
	rtp.R8 = r8
	rtp.R9 = r9
	rtp.R10 = r10
	rtp.R11 = r11
	rtp.R12 = r12
	rtp.R13 = r13
	rtp.R14 = r14
	rtp.R15 = r15
	rtp.R16 = r16
	return
}
