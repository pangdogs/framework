package variant

import (
	"kit.golaxy.org/plugins/util/binaryutil"
)

// Int32 builtin int32
type Int32 int32

// Read implements io.Reader
func (v Int32) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteInt32(int32(v)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (v *Int32) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	val, err := bs.ReadInt32()
	if err != nil {
		return bs.BytesRead(), err
	}
	*v = Int32(val)
	return bs.BytesRead(), nil
}

// Size 大小
func (Int32) Size() int {
	return binaryutil.SizeofInt32()
}

// Type 类型
func (Int32) Type() TypeId {
	return TypeId_Int32
}

// Indirect 原始值
func (v Int32) Indirect() any {
	return int32(v)
}
