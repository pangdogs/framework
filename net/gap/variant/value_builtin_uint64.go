package variant

import (
	"git.golaxy.org/framework/util/binaryutil"
)

// Uint64 builtin uint64
type Uint64 uint64

// Read implements io.Reader
func (v Uint64) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteUvarint(uint64(v)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (v *Uint64) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	val, err := bs.ReadUvarint()
	if err != nil {
		return bs.BytesRead(), err
	}
	*v = Uint64(val)
	return bs.BytesRead(), nil
}

// Size 大小
func (v Uint64) Size() int {
	return binaryutil.SizeofUvarint(uint64(v))
}

// TypeId 类型
func (Uint64) TypeId() TypeId {
	return TypeId_Uint64
}

// Indirect 原始值
func (v Uint64) Indirect() any {
	return uint64(v)
}
