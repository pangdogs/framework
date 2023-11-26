package variant

import (
	"kit.golaxy.org/plugins/util/binaryutil"
)

// Int16 builtin int16
type Int16 int16

// Read implements io.Reader
func (v Int16) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteInt16(int16(v)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (v *Int16) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	val, err := bs.ReadInt16()
	if err != nil {
		return bs.BytesRead(), err
	}
	*v = Int16(val)
	return bs.BytesRead(), nil
}

// Size 大小
func (Int16) Size() int {
	return binaryutil.SizeofInt16()
}

// Type 类型
func (Int16) Type() TypeId {
	return TypeId_Int16
}
