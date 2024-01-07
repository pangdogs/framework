package variant

import (
	"kit.golaxy.org/plugins/util/binaryutil"
)

// Bytes bytes
type Bytes []byte

// Read implements io.Reader
func (v Bytes) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteBytes(v); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (v *Bytes) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	val, err := bs.ReadBytes()
	if err != nil {
		return bs.BytesRead(), err
	}
	*v = val
	return bs.BytesRead(), nil
}

// Size 大小
func (v Bytes) Size() int {
	return binaryutil.SizeofBytes(v)
}

// Type 类型
func (Bytes) Type() TypeId {
	return TypeId_Bytes
}

// Indirect 原始值
func (v Bytes) Indirect() any {
	return []byte(v)
}
