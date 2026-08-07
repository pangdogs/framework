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
	// ErrMethodResultCountMismatch 表示 RPC 返回值数量少于调用方要求。
	ErrMethodResultCountMismatch = errors.New("rpc: method result count mismatch")
	// ErrMethodResultTypeMismatch 表示 RPC 返回值无法转换为调用方指定的类型。
	ErrMethodResultTypeMismatch = errors.New("rpc: method result type mismatch")
)

func parseResult[T any](retArr variant.Array, idx int) (T, error) {
	v := retArr.Items[idx]

	ret, ok := v.Value.Indirect().(T)
	if ok {
		return ret, nil
	}

	retRV, err := v.ToNative(reflect.TypeFor[T]())
	if err != nil {
		return types.Zero[T](), ErrMethodResultTypeMismatch
	}

	switch retRV.Kind() {
	case reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		if retRV.IsNil() {
			return types.Zero[T](), nil
		}
	}

	return retRV.Interface().(T), nil
}

// ParseResults 从异步结果中解析未指定类型的返回值列表。
func ParseResults(ret async.Result) (rvs ResultValues) {
	if !ret.OK() {
		rvs.Error = ret.Error
		return
	}

	if ret.Value == nil {
		return
	}

	retArr, ok := ret.Value.(variant.Array)
	if !ok || len(retArr.Items) < 1 {
		return
	}

	rets := make([]any, len(retArr.Items))

	for i := range rets {
		rets[i] = retArr.Items[i].Value.Indirect()
	}

	rvs.Values = rets
	return
}

// ParseVoid 将异步结果解析为无返回值结果。
func ParseVoid(ret async.Result) (rtp ResultTupleVoid) {
	if !ret.OK() {
		rtp.Error = ret.Error
		return
	}
	return
}

// Parse1 从异步结果中解析至少一个指定类型的返回值，多余项会被忽略。
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
	if !ok || len(retArr.Items) < 1 {
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

// Parse2 从异步结果中解析至少两个指定类型的返回值，多余项会被忽略。
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
	if !ok || len(retArr.Items) < 2 {
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

// Parse3 从异步结果中解析至少三个指定类型的返回值，多余项会被忽略。
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
	if !ok || len(retArr.Items) < 3 {
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

// Parse4 从异步结果中解析至少四个指定类型的返回值，多余项会被忽略。
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
	if !ok || len(retArr.Items) < 4 {
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

// Parse5 从异步结果中解析至少五个指定类型的返回值，多余项会被忽略。
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
	if !ok || len(retArr.Items) < 5 {
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

// Parse6 从异步结果中解析至少六个指定类型的返回值，多余项会被忽略。
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
	if !ok || len(retArr.Items) < 6 {
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

// Parse7 从异步结果中解析至少七个指定类型的返回值，多余项会被忽略。
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
	if !ok || len(retArr.Items) < 7 {
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

// Parse8 从异步结果中解析至少八个指定类型的返回值，多余项会被忽略。
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
	if !ok || len(retArr.Items) < 8 {
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

// Parse9 从异步结果中解析至少九个指定类型的返回值，多余项会被忽略。
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
	if !ok || len(retArr.Items) < 9 {
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

// Parse10 从异步结果中解析至少十个指定类型的返回值，多余项会被忽略。
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
	if !ok || len(retArr.Items) < 10 {
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

// Parse11 从异步结果中解析至少十一个指定类型的返回值，多余项会被忽略。
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
	if !ok || len(retArr.Items) < 11 {
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

// Parse12 从异步结果中解析至少十二个指定类型的返回值，多余项会被忽略。
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
	if !ok || len(retArr.Items) < 12 {
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

// Parse13 从异步结果中解析至少十三个指定类型的返回值，多余项会被忽略。
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
	if !ok || len(retArr.Items) < 13 {
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

// Parse14 从异步结果中解析至少十四个指定类型的返回值，多余项会被忽略。
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
	if !ok || len(retArr.Items) < 14 {
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

// Parse15 从异步结果中解析至少十五个指定类型的返回值，多余项会被忽略。
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
	if !ok || len(retArr.Items) < 15 {
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

// Parse16 从异步结果中解析至少十六个指定类型的返回值，多余项会被忽略。
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
	if !ok || len(retArr.Items) < 16 {
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
