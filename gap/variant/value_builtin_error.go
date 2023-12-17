package variant

import (
	"fmt"
	"kit.golaxy.org/plugins/util/binaryutil"
)

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

func (v Error) Error() string {
	return fmt.Sprintf("(%d) %s", v.Code, v.Message)
}
