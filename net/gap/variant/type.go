package variant

import (
	"git.golaxy.org/framework/util/binaryutil"
	"reflect"
)

// TypeId 类型Id
type TypeId uint32

// Read implements io.Reader
func (t TypeId) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteUvarint(uint64(t)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (t *TypeId) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	v, err := bs.ReadUvarint()
	if err != nil {
		return bs.BytesRead(), err
	}
	*t = TypeId(v)

	return bs.BytesRead(), nil
}

// Size 大小
func (t TypeId) Size() int {
	return binaryutil.SizeofUvarint(uint64(t))
}

// New 创建对象指针
func (t TypeId) New() (Value, error) {
	return variantCreator.New(t)
}

// NewReflected 创建反射对象指针
func (t TypeId) NewReflected() (reflect.Value, error) {
	return variantCreator.NewReflected(t)
}
