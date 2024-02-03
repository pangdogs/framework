package variant

import (
	"errors"
	"fmt"
	"git.golaxy.org/framework/util/binaryutil"
)

func MakeError(err error) *Error {
	if err == nil {
		return &Error{}
	}

	var varErr *Error
	if !errors.As(err, &varErr) {
		return &Error{
			Code:    -1,
			Message: err.Error(),
		}
	}

	return varErr
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

// Type 类型
func (Error) Type() TypeId {
	return TypeId_Error
}

// Indirect 原始值
func (v Error) Indirect() any {
	return v
}

func (v Error) Error() string {
	return fmt.Sprintf("(%d) %s", v.Code, v.Message)
}

func (v Error) OK() bool {
	return v.Code == 0
}
