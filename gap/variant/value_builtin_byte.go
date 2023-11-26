package variant

import (
	"kit.golaxy.org/plugins/util/binaryutil"
)

// Byte builtin byte
type Byte byte

// Read implements io.Reader
func (v Byte) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteByte(byte(v)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (v *Byte) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	val, err := bs.ReadByte()
	if err != nil {
		return bs.BytesRead(), err
	}
	*v = Byte(val)
	return bs.BytesRead(), nil
}

// Size 大小
func (Byte) Size() int {
	return binaryutil.SizeofByte()
}

// Type 类型
func (Byte) Type() TypeId {
	return TypeId_Byte
}
