package variant

import (
	"git.golaxy.org/framework/util/binaryutil"
)

// Int8 builtin int8
type Int8 int8

// Read implements io.Reader
func (v Int8) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteInt8(int8(v)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (v *Int8) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	val, err := bs.ReadInt8()
	if err != nil {
		return bs.BytesRead(), err
	}
	*v = Int8(val)
	return bs.BytesRead(), nil
}

// Size 大小
func (Int8) Size() int {
	return binaryutil.SizeofInt8()
}

// Type 类型
func (Int8) Type() TypeId {
	return TypeId_Int8
}

// Indirect 原始值
func (v Int8) Indirect() any {
	return int8(v)
}
