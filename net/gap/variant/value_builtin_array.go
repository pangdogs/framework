package variant

import (
	"git.golaxy.org/framework/util/binaryutil"
)

// Array array
type Array []Variant

// Read implements io.Reader
func (v Array) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	if err := bs.WriteUvarint(uint64(len(v))); err != nil {
		return bs.BytesWritten(), err
	}

	for i := range v {
		if _, err := binaryutil.ReadFrom(&bs, v[i]); err != nil {
			return bs.BytesWritten(), err
		}
	}

	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (v *Array) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	l, err := bs.ReadUvarint()
	if err != nil {
		return bs.BytesRead(), err
	}

	*v = make([]Variant, l)

	for i := uint64(0); i < l; i++ {
		if _, err := bs.WriteTo(&(*v)[i]); err != nil {
			return bs.BytesRead(), err
		}
	}

	return bs.BytesRead(), nil
}

// Size 大小
func (v Array) Size() int {
	n := binaryutil.SizeofUvarint(uint64(len(v)))
	for i := range v {
		n += v[i].Size()
	}
	return n
}

// TypeId 类型
func (Array) TypeId() TypeId {
	return TypeId_Array
}

// Indirect 原始值
func (v Array) Indirect() any {
	return v
}

// Release 释放资源
func (v Array) Release() {
	for i := range v {
		it := &v[i]

		if it.Readonly() {
			it.ValueReadonly.Release()
		}
	}
}
