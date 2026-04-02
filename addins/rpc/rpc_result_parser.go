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
	"reflect"

	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/types"
	"git.golaxy.org/framework/net/gap/variant"
)

var (
	ErrMethodResultCountMismatch = errors.New("rpc: method result count mismatch")
	ErrMethodResultTypeMismatch  = errors.New("rpc: method result type mismatch")
)

func parseResult[T any](retArr variant.Array, idx int) (T, error) {
	v := retArr[idx]

	ret, ok := v.Value.Indirect().(T)
	if ok {
		return ret, nil
	}

	retRV, err := v.Convert(reflect.TypeFor[T]())
	if err != nil {
		return types.Zero[T](), ErrMethodResultTypeMismatch
	}

	if retRV.IsNil() {
		return types.Zero[T](), nil
	}

	return retRV.Interface().(T), nil
}

func ParseResults(ret async.Result) (rvs ResultValues) {
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

func ParseVoid(ret async.Result) (rtp ResultTuple0) {
	if !ret.OK() {
		rtp.Error = ret.Error
		return
	}
	return
}

func Parse1[T1 any](ret async.Result) (rtp ResultTuple1[T1]) {
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

	r1, err := parseResult[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	rtp.R1 = r1
	return
}

func Parse2[T1 any, T2 any](ret async.Result) (rtp ResultTuple2[T1, T2]) {
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

	r1, err := parseResult[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseResult[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	rtp.R1 = r1
	rtp.R2 = r2
	return
}

func Parse3[T1 any, T2 any, T3 any](ret async.Result) (rtp ResultTuple3[T1, T2, T3]) {
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

	r1, err := parseResult[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseResult[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseResult[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	rtp.R1 = r1
	rtp.R2 = r2
	rtp.R3 = r3
	return
}

func Parse4[T1 any, T2 any, T3 any, T4 any](ret async.Result) (rtp ResultTuple4[T1, T2, T3, T4]) {
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

	r1, err := parseResult[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseResult[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseResult[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	r4, err := parseResult[T4](retArr, 3)
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

func Parse5[T1 any, T2 any, T3 any, T4 any, T5 any](ret async.Result) (rtp ResultTuple5[T1, T2, T3, T4, T5]) {
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

	r1, err := parseResult[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseResult[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseResult[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	r4, err := parseResult[T4](retArr, 3)
	if err != nil {
		rtp.Error = err
		return
	}

	r5, err := parseResult[T5](retArr, 4)
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

func Parse6[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any](ret async.Result) (rtp ResultTuple6[T1, T2, T3, T4, T5, T6]) {
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

	r1, err := parseResult[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseResult[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseResult[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	r4, err := parseResult[T4](retArr, 3)
	if err != nil {
		rtp.Error = err
		return
	}

	r5, err := parseResult[T5](retArr, 4)
	if err != nil {
		rtp.Error = err
		return
	}

	r6, err := parseResult[T6](retArr, 5)
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

func Parse7[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any](ret async.Result) (rtp ResultTuple7[T1, T2, T3, T4, T5, T6, T7]) {
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

	r1, err := parseResult[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseResult[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseResult[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	r4, err := parseResult[T4](retArr, 3)
	if err != nil {
		rtp.Error = err
		return
	}

	r5, err := parseResult[T5](retArr, 4)
	if err != nil {
		rtp.Error = err
		return
	}

	r6, err := parseResult[T6](retArr, 5)
	if err != nil {
		rtp.Error = err
		return
	}

	r7, err := parseResult[T7](retArr, 6)
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

func Parse8[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any](ret async.Result) (rtp ResultTuple8[T1, T2, T3, T4, T5, T6, T7, T8]) {
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

	r1, err := parseResult[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseResult[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseResult[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	r4, err := parseResult[T4](retArr, 3)
	if err != nil {
		rtp.Error = err
		return
	}

	r5, err := parseResult[T5](retArr, 4)
	if err != nil {
		rtp.Error = err
		return
	}

	r6, err := parseResult[T6](retArr, 5)
	if err != nil {
		rtp.Error = err
		return
	}

	r7, err := parseResult[T7](retArr, 6)
	if err != nil {
		rtp.Error = err
		return
	}

	r8, err := parseResult[T8](retArr, 7)
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

func Parse9[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any](ret async.Result) (rtp ResultTuple9[T1, T2, T3, T4, T5, T6, T7, T8, T9]) {
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

	r1, err := parseResult[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseResult[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseResult[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	r4, err := parseResult[T4](retArr, 3)
	if err != nil {
		rtp.Error = err
		return
	}

	r5, err := parseResult[T5](retArr, 4)
	if err != nil {
		rtp.Error = err
		return
	}

	r6, err := parseResult[T6](retArr, 5)
	if err != nil {
		rtp.Error = err
		return
	}

	r7, err := parseResult[T7](retArr, 6)
	if err != nil {
		rtp.Error = err
		return
	}

	r8, err := parseResult[T8](retArr, 7)
	if err != nil {
		rtp.Error = err
		return
	}

	r9, err := parseResult[T9](retArr, 8)
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

func Parse10[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any](ret async.Result) (rtp ResultTuple10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) {
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

	r1, err := parseResult[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseResult[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseResult[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	r4, err := parseResult[T4](retArr, 3)
	if err != nil {
		rtp.Error = err
		return
	}

	r5, err := parseResult[T5](retArr, 4)
	if err != nil {
		rtp.Error = err
		return
	}

	r6, err := parseResult[T6](retArr, 5)
	if err != nil {
		rtp.Error = err
		return
	}

	r7, err := parseResult[T7](retArr, 6)
	if err != nil {
		rtp.Error = err
		return
	}

	r8, err := parseResult[T8](retArr, 7)
	if err != nil {
		rtp.Error = err
		return
	}

	r9, err := parseResult[T9](retArr, 8)
	if err != nil {
		rtp.Error = err
		return
	}

	r10, err := parseResult[T10](retArr, 9)
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

func Parse11[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any](ret async.Result) (rtp ResultTuple11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]) {
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

	r1, err := parseResult[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseResult[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseResult[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	r4, err := parseResult[T4](retArr, 3)
	if err != nil {
		rtp.Error = err
		return
	}

	r5, err := parseResult[T5](retArr, 4)
	if err != nil {
		rtp.Error = err
		return
	}

	r6, err := parseResult[T6](retArr, 5)
	if err != nil {
		rtp.Error = err
		return
	}

	r7, err := parseResult[T7](retArr, 6)
	if err != nil {
		rtp.Error = err
		return
	}

	r8, err := parseResult[T8](retArr, 7)
	if err != nil {
		rtp.Error = err
		return
	}

	r9, err := parseResult[T9](retArr, 8)
	if err != nil {
		rtp.Error = err
		return
	}

	r10, err := parseResult[T10](retArr, 9)
	if err != nil {
		rtp.Error = err
		return
	}

	r11, err := parseResult[T11](retArr, 10)
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

func Parse12[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any](ret async.Result) (rtp ResultTuple12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) {
	if !ret.OK() {
		rtp.Error = ret.Error
		return
	}

	if ret.Value == nil {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	retArr, ok := ret.Value.(variant.Array)
	if !ok || len(retArr) < 12 {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	r1, err := parseResult[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseResult[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseResult[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	r4, err := parseResult[T4](retArr, 3)
	if err != nil {
		rtp.Error = err
		return
	}

	r5, err := parseResult[T5](retArr, 4)
	if err != nil {
		rtp.Error = err
		return
	}

	r6, err := parseResult[T6](retArr, 5)
	if err != nil {
		rtp.Error = err
		return
	}

	r7, err := parseResult[T7](retArr, 6)
	if err != nil {
		rtp.Error = err
		return
	}

	r8, err := parseResult[T8](retArr, 7)
	if err != nil {
		rtp.Error = err
		return
	}

	r9, err := parseResult[T9](retArr, 8)
	if err != nil {
		rtp.Error = err
		return
	}

	r10, err := parseResult[T10](retArr, 9)
	if err != nil {
		rtp.Error = err
		return
	}

	r11, err := parseResult[T11](retArr, 10)
	if err != nil {
		rtp.Error = err
		return
	}

	r12, err := parseResult[T12](retArr, 11)
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

func Parse13[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any](ret async.Result) (rtp ResultTuple13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) {
	if !ret.OK() {
		rtp.Error = ret.Error
		return
	}

	if ret.Value == nil {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	retArr, ok := ret.Value.(variant.Array)
	if !ok || len(retArr) < 13 {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	r1, err := parseResult[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseResult[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseResult[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	r4, err := parseResult[T4](retArr, 3)
	if err != nil {
		rtp.Error = err
		return
	}

	r5, err := parseResult[T5](retArr, 4)
	if err != nil {
		rtp.Error = err
		return
	}

	r6, err := parseResult[T6](retArr, 5)
	if err != nil {
		rtp.Error = err
		return
	}

	r7, err := parseResult[T7](retArr, 6)
	if err != nil {
		rtp.Error = err
		return
	}

	r8, err := parseResult[T8](retArr, 7)
	if err != nil {
		rtp.Error = err
		return
	}

	r9, err := parseResult[T9](retArr, 8)
	if err != nil {
		rtp.Error = err
		return
	}

	r10, err := parseResult[T10](retArr, 9)
	if err != nil {
		rtp.Error = err
		return
	}

	r11, err := parseResult[T11](retArr, 10)
	if err != nil {
		rtp.Error = err
		return
	}

	r12, err := parseResult[T12](retArr, 11)
	if err != nil {
		rtp.Error = err
		return
	}

	r13, err := parseResult[T13](retArr, 12)
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

func Parse14[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any](ret async.Result) (rtp ResultTuple14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) {
	if !ret.OK() {
		rtp.Error = ret.Error
		return
	}

	if ret.Value == nil {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	retArr, ok := ret.Value.(variant.Array)
	if !ok || len(retArr) < 14 {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	r1, err := parseResult[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseResult[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseResult[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	r4, err := parseResult[T4](retArr, 3)
	if err != nil {
		rtp.Error = err
		return
	}

	r5, err := parseResult[T5](retArr, 4)
	if err != nil {
		rtp.Error = err
		return
	}

	r6, err := parseResult[T6](retArr, 5)
	if err != nil {
		rtp.Error = err
		return
	}

	r7, err := parseResult[T7](retArr, 6)
	if err != nil {
		rtp.Error = err
		return
	}

	r8, err := parseResult[T8](retArr, 7)
	if err != nil {
		rtp.Error = err
		return
	}

	r9, err := parseResult[T9](retArr, 8)
	if err != nil {
		rtp.Error = err
		return
	}

	r10, err := parseResult[T10](retArr, 9)
	if err != nil {
		rtp.Error = err
		return
	}

	r11, err := parseResult[T11](retArr, 10)
	if err != nil {
		rtp.Error = err
		return
	}

	r12, err := parseResult[T12](retArr, 11)
	if err != nil {
		rtp.Error = err
		return
	}

	r13, err := parseResult[T13](retArr, 12)
	if err != nil {
		rtp.Error = err
		return
	}

	r14, err := parseResult[T14](retArr, 13)
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

func Parse15[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any](ret async.Result) (rtp ResultTuple15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) {
	if !ret.OK() {
		rtp.Error = ret.Error
		return
	}

	if ret.Value == nil {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	retArr, ok := ret.Value.(variant.Array)
	if !ok || len(retArr) < 15 {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	r1, err := parseResult[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseResult[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseResult[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	r4, err := parseResult[T4](retArr, 3)
	if err != nil {
		rtp.Error = err
		return
	}

	r5, err := parseResult[T5](retArr, 4)
	if err != nil {
		rtp.Error = err
		return
	}

	r6, err := parseResult[T6](retArr, 5)
	if err != nil {
		rtp.Error = err
		return
	}

	r7, err := parseResult[T7](retArr, 6)
	if err != nil {
		rtp.Error = err
		return
	}

	r8, err := parseResult[T8](retArr, 7)
	if err != nil {
		rtp.Error = err
		return
	}

	r9, err := parseResult[T9](retArr, 8)
	if err != nil {
		rtp.Error = err
		return
	}

	r10, err := parseResult[T10](retArr, 9)
	if err != nil {
		rtp.Error = err
		return
	}

	r11, err := parseResult[T11](retArr, 10)
	if err != nil {
		rtp.Error = err
		return
	}

	r12, err := parseResult[T12](retArr, 11)
	if err != nil {
		rtp.Error = err
		return
	}

	r13, err := parseResult[T13](retArr, 12)
	if err != nil {
		rtp.Error = err
		return
	}

	r14, err := parseResult[T14](retArr, 13)
	if err != nil {
		rtp.Error = err
		return
	}

	r15, err := parseResult[T15](retArr, 14)
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

func Parse16[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any](ret async.Result) (rtp ResultTuple16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16]) {
	if !ret.OK() {
		rtp.Error = ret.Error
		return
	}

	if ret.Value == nil {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	retArr, ok := ret.Value.(variant.Array)
	if !ok || len(retArr) < 16 {
		rtp.Error = ErrMethodResultCountMismatch
		return
	}

	r1, err := parseResult[T1](retArr, 0)
	if err != nil {
		rtp.Error = err
		return
	}

	r2, err := parseResult[T2](retArr, 1)
	if err != nil {
		rtp.Error = err
		return
	}

	r3, err := parseResult[T3](retArr, 2)
	if err != nil {
		rtp.Error = err
		return
	}

	r4, err := parseResult[T4](retArr, 3)
	if err != nil {
		rtp.Error = err
		return
	}

	r5, err := parseResult[T5](retArr, 4)
	if err != nil {
		rtp.Error = err
		return
	}

	r6, err := parseResult[T6](retArr, 5)
	if err != nil {
		rtp.Error = err
		return
	}

	r7, err := parseResult[T7](retArr, 6)
	if err != nil {
		rtp.Error = err
		return
	}

	r8, err := parseResult[T8](retArr, 7)
	if err != nil {
		rtp.Error = err
		return
	}

	r9, err := parseResult[T9](retArr, 8)
	if err != nil {
		rtp.Error = err
		return
	}

	r10, err := parseResult[T10](retArr, 9)
	if err != nil {
		rtp.Error = err
		return
	}

	r11, err := parseResult[T11](retArr, 10)
	if err != nil {
		rtp.Error = err
		return
	}

	r12, err := parseResult[T12](retArr, 11)
	if err != nil {
		rtp.Error = err
		return
	}

	r13, err := parseResult[T13](retArr, 12)
	if err != nil {
		rtp.Error = err
		return
	}

	r14, err := parseResult[T14](retArr, 13)
	if err != nil {
		rtp.Error = err
		return
	}

	r15, err := parseResult[T15](retArr, 14)
	if err != nil {
		rtp.Error = err
		return
	}

	r16, err := parseResult[T16](retArr, 15)
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
