package variant

import (
	"git.golaxy.org/framework/util/binaryutil"
)

// Float builtin float
type Float float32

// Read implements io.Reader
func (v Float) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteFloat(float32(v)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (v *Float) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	val, err := bs.ReadFloat()
	if err != nil {
		return bs.BytesRead(), err
	}
	*v = Float(val)
	return bs.BytesRead(), nil
}

// Size 大小
func (Float) Size() int {
	return binaryutil.SizeofFloat()
}

// TypeId 类型
func (Float) TypeId() TypeId {
	return TypeId_Float
}

// Indirect 原始值
func (v Float) Indirect() any {
	return float32(v)
}
