package variant

import (
	"kit.golaxy.org/plugins/util/binaryutil"
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

// Type 类型
func (Uint) Type() TypeId {
	return TypeId_Uint
}
