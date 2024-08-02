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

package variant

import (
	"errors"
	"fmt"
	"git.golaxy.org/framework/utils/binaryutil"
)

func MakeError(err error) *Error {
	if err == nil {
		return &Error{}
	}

	var varErr *Error
	if !errors.As(err, &varErr) {
		return Errorln(-1, err.Error())
	}

	return varErr
}

func Errorf(code int32, format string, args ...any) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

func Errorln(code int32, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Error builtin error
type Error struct {
	Code    int32
	Message string
}

// Read implements io.Reader
func (v Error) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteInt32(v.Code); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteString(v.Message); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (v *Error) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	var err error

	v.Code, err = bs.ReadInt32()
	if err != nil {
		return bs.BytesRead(), err
	}

	v.Message, err = bs.ReadString()
	if err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 大小
func (v Error) Size() int {
	return binaryutil.SizeofInt32() + binaryutil.SizeofString(v.Message)
}

// TypeId 类型
func (Error) TypeId() TypeId {
	return TypeId_Error
}

// Indirect 原始值
func (v *Error) Indirect() any {
	return v
}

func (v Error) Error() string {
	return fmt.Sprintf("(%d) %s", v.Code, v.Message)
}

func (v Error) OK() bool {
	return v.Code == 0
}
