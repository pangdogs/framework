package variant

import (
	"git.golaxy.org/framework/utils/binaryutil"
)

// Int64 builtin int64
type Int64 int64

// Read implements io.Reader
func (v Int64) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteVarint(int64(v)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (v *Int64) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	val, err := bs.ReadVarint()
	if err != nil {
		return bs.BytesRead(), err
	}
	*v = Int64(val)
	return bs.BytesRead(), nil
}

// Size 大小
func (v Int64) Size() int {
	return binaryutil.SizeofVarint(int64(v))
}

// TypeId 类型
func (Int64) TypeId() TypeId {
	return TypeId_Int64
}

// Indirect 原始值
func (v Int64) Indirect() any {
	return int64(v)
}
