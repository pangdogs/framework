package variant

import (
	"kit.golaxy.org/plugins/util/binaryutil"
)

// Uint16 builtin uint16
type Uint16 uint16

// Read implements io.Reader
func (v Uint16) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteUint16(uint16(v)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (v *Uint16) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	val, err := bs.ReadUint16()
	if err != nil {
		return bs.BytesRead(), err
	}
	*v = Uint16(val)
	return bs.BytesRead(), nil
}

// Size 大小
func (Uint16) Size() int {
	return binaryutil.SizeofUint16()
}

// Type 类型
func (Uint16) Type() TypeId {
	return TypeId_Uint16
}
