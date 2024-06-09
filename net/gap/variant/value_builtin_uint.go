package variant

import (
	"git.golaxy.org/framework/util/binaryutil"
)

// Uint builtin uint
type Uint uint

// Read implements io.Reader
func (v Uint) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteUvarint(uint64(v)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (v *Uint) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	val, err := bs.ReadUvarint()
	if err != nil {
		return bs.BytesRead(), err
	}
	*v = Uint(val)
	return bs.BytesRead(), nil
}

// Size 大小
func (v Uint) Size() int {
	return binaryutil.SizeofUvarint(uint64(v))
}

// TypeId 类型
func (Uint) TypeId() TypeId {
	return TypeId_Uint
}

// Indirect 原始值
func (v Uint) Indirect() any {
	return uint(v)
}

// Release 释放资源
func (Uint) Release() {}
