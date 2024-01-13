package variant

import (
	"git.golaxy.org/plugins/util/binaryutil"
)

// Uint8 builtin uint8
type Uint8 uint8

// Read implements io.Reader
func (v Uint8) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteUint8(uint8(v)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (v *Uint8) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	val, err := bs.ReadUint8()
	if err != nil {
		return bs.BytesRead(), err
	}
	*v = Uint8(val)
	return bs.BytesRead(), nil
}

// Size 大小
func (Uint8) Size() int {
	return binaryutil.SizeofUint8()
}

// Type 类型
func (Uint8) Type() TypeId {
	return TypeId_Uint8
}

// Indirect 原始值
func (v Uint8) Indirect() any {
	return uint8(v)
}
