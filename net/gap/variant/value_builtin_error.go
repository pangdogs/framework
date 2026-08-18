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
	"io"

	"git.golaxy.org/framework/utils/binaryutil"
)

// NewError 将普通 error 转换为可传输错误；nil 转换为成功结果。
func NewError(err error) *Error {
	if err == nil {
		return &Error{}
	}

	var varErr *Error
	if !errors.As(err, &varErr) {
		return Errorln(-1, err.Error())
	}

	return varErr
}

// Errorf 创建带错误码和格式化消息的可传输错误。
func Errorf(code int32, format string, args ...any) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// Errorln 创建带错误码和原始消息的可传输错误。
func Errorln(code int32, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Error 是包含数值错误码和消息的可传输错误。
type Error struct {
	Code    int32  // 错误码；零表示成功。
	Message string // 错误消息。
}

// Read 将错误编码到 p。
func (v Error) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteInt32(v.Code); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteString(v.Message); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write 从 p 解码错误。
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

// Size 返回错误编码后的字节数。
func (v Error) Size() int {
	return binaryutil.SizeofInt32 + binaryutil.SizeofString(v.Message)
}

// TypeID 返回错误的内置类型 ID。
func (Error) TypeID() TypeID {
	return TypeID_Error
}

// Indirect 返回错误指针。
func (v *Error) Indirect() any {
	return v
}

// Error 返回包含错误码和消息的文本。
func (v Error) Error() string {
	return fmt.Sprintf("(%d) %s", v.Code, v.Message)
}

// OK 报告错误码是否为零。
func (v Error) OK() bool {
	return v.Code == 0
}
