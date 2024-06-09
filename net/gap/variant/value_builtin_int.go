package variant

import (
	"git.golaxy.org/framework/util/binaryutil"
)

// Int builtin int
type Int int

// Read implements io.Reader
func (v Int) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteVarint(int64(v)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (v *Int) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	val, err := bs.ReadVarint()
	if err != nil {
		return bs.BytesRead(), err
	}
	*v = Int(val)
	return bs.BytesRead(), nil
}

// Size 大小
func (v Int) Size() int {
	return binaryutil.SizeofVarint(int64(v))
}

// TypeId 类型
func (Int) TypeId() TypeId {
	return TypeId_Int
}

// Indirect 原始值
func (v Int) Indirect() any {
	return int(v)
}

// Release 释放资源
func (Int) Release() {}
