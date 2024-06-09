package variant

import (
	"git.golaxy.org/framework/util/binaryutil"
)

// Bool builtin bool
type Bool bool

// Read implements io.Reader
func (v Bool) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteBool(bool(v)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (v *Bool) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	val, err := bs.ReadBool()
	if err != nil {
		return bs.BytesRead(), err
	}
	*v = Bool(val)
	return bs.BytesRead(), nil
}

// Size 大小
func (Bool) Size() int {
	return binaryutil.SizeofBool()
}

// TypeId 类型
func (Bool) TypeId() TypeId {
	return TypeId_Bool
}

// Indirect 原始值
func (v Bool) Indirect() any {
	return bool(v)
}

// Release 释放资源
func (Bool) Release() {}
