package variant

import (
	"kit.golaxy.org/plugins/util/binaryutil"
)

// Uint32 builtin uint32
type Uint32 uint32

// Read implements io.Reader
func (v Uint32) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteUint32(uint32(v)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (v *Uint32) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	val, err := bs.ReadUint32()
	if err != nil {
		return bs.BytesRead(), err
	}
	*v = Uint32(val)
	return bs.BytesRead(), nil
}

// Size 大小
func (Uint32) Size() int {
	return binaryutil.SizeofUint32()
}

// Type 类型
func (Uint32) Type() TypeId {
	return TypeId_Uint32
}
