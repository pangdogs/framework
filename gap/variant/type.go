package variant

import (
	"kit.golaxy.org/plugins/util/binaryutil"
)

// TypeId 类型Id
type TypeId uint32

// Read implements io.Reader
func (t TypeId) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteUint32(uint32(t)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (t *TypeId) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	v, err := bs.ReadUint32()
	if err != nil {
		return bs.BytesRead(), err
	}
	*t = TypeId(v)
	return bs.BytesRead(), nil
}

// Size 大小
func (TypeId) Size() int {
	return binaryutil.SizeofUint32()
}

// New 创建值
func (t TypeId) New() (Value, error) {
	return variantCreator.Spawn(t)
}
